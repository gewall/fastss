package capture

import (
	"fmt"
	"image"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/kbinani/screenshot"
	"golang.org/x/sys/windows"
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	gdi32    = windows.NewLazySystemDLL("gdi32.dll")
	dwmapi   = windows.NewLazySystemDLL("dwmapi.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procGetForegroundWindow           = user32.NewProc("GetForegroundWindow")
	procGetWindowRect                 = user32.NewProc("GetWindowRect")
	procGetWindowTextW                = user32.NewProc("GetWindowTextW")
	procIsWindowVisible               = user32.NewProc("IsWindowVisible")
	procIsIconic                      = user32.NewProc("IsIconic")
	procEnumWindows                   = user32.NewProc("EnumWindows")
	procSetProcessDpiAwarenessContext = user32.NewProc("SetProcessDpiAwarenessContext")
	procSetProcessDPIAware            = user32.NewProc("SetProcessDPIAware")
	procGetWindowThreadProcessId      = user32.NewProc("GetWindowThreadProcessId")
	procSetForegroundWindow           = user32.NewProc("SetForegroundWindow")
	procShowWindow                    = user32.NewProc("ShowWindow")
	procBringWindowToTop              = user32.NewProc("BringWindowToTop")
	procSwitchToThisWindow            = user32.NewProc("SwitchToThisWindow")
	procAttachThreadInput             = user32.NewProc("AttachThreadInput")
	procGetWindow                     = user32.NewProc("GetWindow")
	procGetDC                         = user32.NewProc("GetDC")
	procReleaseDC                     = user32.NewProc("ReleaseDC")
	procPrintWindow                   = user32.NewProc("PrintWindow")

	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procDeleteDC               = gdi32.NewProc("DeleteDC")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procGetDIBits              = gdi32.NewProc("GetDIBits")

	procGetCurrentThreadId = kernel32.NewProc("GetCurrentThreadId")
	procGetConsoleWindow   = kernel32.NewProc("GetConsoleWindow")

	procDwmGetWindowAttribute = dwmapi.NewProc("DwmGetWindowAttribute")
)

const (
	DWMWA_EXTENDED_FRAME_BOUNDS = 9
	SW_RESTORE                  = 9
	SW_SHOW                     = 5
	SW_MINIMIZE                 = 6
	SW_HIDE                     = 0
	GW_HWNDNEXT                 = 2
	PW_RENDERFULLCONTENT        = 2
	BI_RGB                      = 0
)

type RECT struct {
	Left, Top, Right, Bottom int32
}

type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

type BITMAPINFO struct {
	BmiHeader BITMAPINFOHEADER
	BmiColors [1]uint32
}

type WindowInfo struct {
	Handle    windows.HWND
	Title     string
	Bounds    image.Rectangle
	ProcessID uint32
}

func init() {
	if procSetProcessDpiAwarenessContext.Find() == nil {
		_, _, _ = procSetProcessDpiAwarenessContext.Call(uintptr(unsafe.Pointer(uintptr(0xFFFFFFFFFFFFFFFC))))
	} else if procSetProcessDPIAware.Find() == nil {
		_, _, _ = procSetProcessDPIAware.Call()
	}
}

// GetConsoleHwnd returns the handle of the current terminal/cmd window
func GetConsoleHwnd() uintptr {
	if procGetConsoleWindow.Find() == nil {
		h, _, _ := procGetConsoleWindow.Call()
		return h
	}
	return 0
}

// ForceForegroundWindow forcefully brings the target window to the front and gives it focus
func ForceForegroundWindow(hwnd uintptr) {
	if hwnd == 0 {
		return
	}

	curThreadId, _, _ := procGetCurrentThreadId.Call()
	var targetThreadId uint32
	_, _, _ = procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&targetThreadId)))

	var foreThreadId uint32
	foreHwnd, _, _ := procGetForegroundWindow.Call()
	if foreHwnd != 0 {
		_, _, _ = procGetWindowThreadProcessId.Call(foreHwnd, uintptr(unsafe.Pointer(&foreThreadId)))
	}

	// Attach input threads to bypass Windows foreground restrictions
	if foreThreadId != 0 && foreThreadId != uint32(curThreadId) {
		procAttachThreadInput.Call(uintptr(foreThreadId), curThreadId, 1)
		defer procAttachThreadInput.Call(uintptr(foreThreadId), curThreadId, 0)
	}
	if targetThreadId != 0 && targetThreadId != uint32(curThreadId) {
		procAttachThreadInput.Call(uintptr(targetThreadId), curThreadId, 1)
		defer procAttachThreadInput.Call(uintptr(targetThreadId), curThreadId, 0)
	}

	procShowWindow.Call(hwnd, SW_RESTORE)
	procBringWindowToTop.Call(hwnd)
	procSetForegroundWindow.Call(hwnd)
	if procSwitchToThisWindow.Find() == nil {
		procSwitchToThisWindow.Call(hwnd, 1)
	}
}

// GetWindowTitle returns the window title string for a HWND
func GetWindowTitle(hwnd uintptr) string {
	buf := make([]uint16, 512)
	len, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 512)
	if len == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:len])
}

// GetWindowRectBounds retrieves the real window bounding rectangle including DWM frame
func GetWindowRectBounds(hwnd uintptr) (image.Rectangle, error) {
	var rect RECT
	if procDwmGetWindowAttribute.Find() == nil {
		hr, _, _ := procDwmGetWindowAttribute.Call(
			hwnd,
			uintptr(DWMWA_EXTENDED_FRAME_BOUNDS),
			uintptr(unsafe.Pointer(&rect)),
			uintptr(unsafe.Sizeof(rect)),
		)
		if hr == 0 {
			return image.Rect(int(rect.Left), int(rect.Top), int(rect.Right), int(rect.Bottom)), nil
		}
	}

	ret, _, err := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
	if ret == 0 {
		return image.Rectangle{}, fmt.Errorf("GetWindowRect failed: %w", err)
	}
	return image.Rect(int(rect.Left), int(rect.Top), int(rect.Right), int(rect.Bottom)), nil
}

// isIgnoredTitle checks if a window is an internal OS or CMD shell window
func isIgnoredTitle(title string, hwnd uintptr, consoleHwnd uintptr) bool {
	title = strings.TrimSpace(title)
	if title == "" {
		return true
	}
	if hwnd == consoleHwnd {
		return true
	}

	lower := strings.ToLower(title)
	ignored := []string{
		"program manager",
		"windows input experience",
		"windows shell experience host",
		"settings",
		"microsoft text input application",
	}
	for _, ign := range ignored {
		if lower == ign {
			return true
		}
	}
	return false
}

// ListWindows returns all visible application windows excluding console and OS background windows
func ListWindows() ([]WindowInfo, error) {
	var windowsList []WindowInfo
	consoleHwnd := GetConsoleHwnd()

	cb := syscall.NewCallback(func(hwnd uintptr, lParam uintptr) uintptr {
		vis, _, _ := procIsWindowVisible.Call(hwnd)
		if vis == 0 {
			return 1
		}

		iconic, _, _ := procIsIconic.Call(hwnd)
		if iconic != 0 {
			return 1
		}

		title := GetWindowTitle(hwnd)
		if isIgnoredTitle(title, hwnd, consoleHwnd) {
			return 1
		}

		bounds, err := GetWindowRectBounds(hwnd)
		if err != nil || bounds.Dx() <= 10 || bounds.Dy() <= 10 {
			return 1
		}

		var pid uint32
		_, _, _ = procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))

		windowsList = append(windowsList, WindowInfo{
			Handle:    windows.HWND(hwnd),
			Title:     title,
			Bounds:    bounds,
			ProcessID: pid,
		})

		return 1
	})

	_, _, _ = procEnumWindows.Call(cb, 0)
	return windowsList, nil
}

// FindWindow finds a window matching query (case-insensitive substring match) or active app window
func FindWindow(query string) (*WindowInfo, error) {
	query = strings.TrimSpace(query)
	consoleHwnd := GetConsoleHwnd()

	// If query is active or empty, find foreground window, skipping console if CMD is currently focused
	if strings.EqualFold(query, "active") || strings.EqualFold(query, "current") || query == "" {
		hwnd, _, _ := procGetForegroundWindow.Call()
		if hwnd == consoleHwnd || hwnd == 0 {
			// Find topmost application window beneath console
			list, err := ListWindows()
			if err == nil && len(list) > 0 {
				return &list[0], nil
			}
		}

		if hwnd != 0 {
			title := GetWindowTitle(hwnd)
			bounds, err := GetWindowRectBounds(hwnd)
			if err != nil {
				return nil, err
			}
			var pid uint32
			_, _, _ = procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
			return &WindowInfo{
				Handle:    windows.HWND(hwnd),
				Title:     title,
				Bounds:    bounds,
				ProcessID: pid,
			}, nil
		}
		return nil, fmt.Errorf("no active foreground application found")
	}

	list, err := ListWindows()
	if err != nil {
		return nil, err
	}

	qLower := strings.ToLower(query)
	// Try substring match on Title
	for _, w := range list {
		if strings.Contains(strings.ToLower(w.Title), qLower) {
			return &w, nil
		}
	}

	return nil, fmt.Errorf("window matching '%s' not found. Available open windows:\n%s", query, formatWindowNames(list))
}

func formatWindowNames(list []WindowInfo) string {
	var lines []string
	for _, w := range list {
		lines = append(lines, fmt.Sprintf("  • %s", w.Title))
	}
	if len(lines) == 0 {
		return "  (No visible application windows detected)"
	}
	return strings.Join(lines, "\n")
}

// GetVirtualScreenBounds returns the union of all active display bounds
func GetVirtualScreenBounds() image.Rectangle {
	n := screenshot.NumActiveDisplays()
	if n == 0 {
		return image.Rect(0, 0, 1920, 1080)
	}

	totalRect := screenshot.GetDisplayBounds(0)
	for i := 1; i < n; i++ {
		totalRect = totalRect.Union(screenshot.GetDisplayBounds(i))
	}
	return totalRect
}

// CaptureScreen captures the virtual screen (all monitors combined)
func CaptureScreen() (image.Image, error) {
	bounds := GetVirtualScreenBounds()
	return screenshot.CaptureRect(bounds)
}

// CaptureWindow captures the given window image. It brings the window to front and captures it.
func CaptureWindow(w *WindowInfo, bringToFront bool) (image.Image, error) {
	if bringToFront && w.Handle != 0 {
		ForceForegroundWindow(uintptr(w.Handle))
		time.Sleep(300 * time.Millisecond)

		if b, err := GetWindowRectBounds(uintptr(w.Handle)); err == nil && b.Dx() > 0 && b.Dy() > 0 {
			w.Bounds = b
		}
	}

	// Capture using screen coordinates of window bounding box
	return screenshot.CaptureRect(w.Bounds)
}
