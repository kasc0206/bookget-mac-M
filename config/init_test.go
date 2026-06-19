package config

import (
	"testing"
)

func TestPageRange_Unset(t *testing.T) {
	// 未设置页码范围时，始终返回 true
	Conf.SeqStart = 0
	Conf.SeqEnd = 0

	if !PageRange(0, 10) {
		t.Error("expected true when no range set (index=0)")
	}
	if !PageRange(5, 10) {
		t.Error("expected true when no range set (index=5)")
	}
	if !PageRange(9, 10) {
		t.Error("expected true when no range set (index=9)")
	}
}

func TestPageRange_PositiveRange(t *testing.T) {
	// 设置页码范围 3:7 (从第3页到第7页，1-indexed)
	Conf.SeqStart = 3
	Conf.SeqEnd = 7

	tests := []struct {
		index    int
		size     int
		expected bool
		name     string
	}{
		{0, 10, false, "index=0 (page 1), before start"},
		{1, 10, false, "index=1 (page 2), before start"},
		{2, 10, true, "index=2 (page 3), at start"},
		{3, 10, true, "index=3 (page 4), in range"},
		{5, 10, true, "index=5 (page 6), in range"},
		{6, 10, true, "index=6 (page 7), at end (inclusive)"},
		{7, 10, false, "index=7 (page 8), after end"},
		{9, 10, false, "index=9 (page 10), after end"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PageRange(tt.index, tt.size); got != tt.expected {
				t.Errorf("PageRange(%d, %d) = %v, want %v", tt.index, tt.size, got, tt.expected)
			}
		})
	}
}

func TestPageRange_NegativeEnd(t *testing.T) {
	// 结束页为负数表示从末尾开始计算
	Conf.SeqStart = 5
	Conf.SeqEnd = -2 // 从第5页到倒数第2页
	// 逻辑: index - size >= SeqEnd → index - 10 >= -2 → index >= 8 时返回false
	// 所以 index=7 (page 8) 是最后一个有效页

	tests := []struct {
		index    int
		size     int
		expected bool
		name     string
	}{
		{0, 10, false, "index=0 (page 1), before start"},
		{4, 10, true, "index=4 (page 5), at start"},
		{5, 10, true, "index=5 (page 6), in range"},
		{7, 10, true, "index=7 (page 8), before negative end"},
		{8, 10, false, "index=8 (page 9), at negative end"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PageRange(tt.index, tt.size); got != tt.expected {
				t.Errorf("PageRange(%d, %d) = %v, want %v", tt.index, tt.size, got, tt.expected)
			}
		})
	}
}

func TestPageRange_SinglePage(t *testing.T) {
	// 只有起始页，没有结束页
	Conf.SeqStart = 4
	Conf.SeqEnd = 0

	tests := []struct {
		index    int
		size     int
		expected bool
		name     string
	}{
		{0, 10, false, "index=0 (page 1), before start"},
		{3, 10, true, "index=3 (page 4), at start"},
		{5, 10, true, "index=5 (page 6), after start"},
		{9, 10, true, "index=9 (page 10), after start"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PageRange(tt.index, tt.size); got != tt.expected {
				t.Errorf("PageRange(%d, %d) = %v, want %v", tt.index, tt.size, got, tt.expected)
			}
		})
	}
}

func TestVolumeRange_Unset(t *testing.T) {
	// 未设置册范围时，始终返回 true
	Conf.VolStart = 0
	Conf.VolEnd = 0

	if !VolumeRange(0) {
		t.Error("expected true when no volume range set (index=0)")
	}
	if !VolumeRange(5) {
		t.Error("expected true when no volume range set (index=5)")
	}
}

func TestVolumeRange_PositiveRange(t *testing.T) {
	// 设置册范围 2:5 (第2册到第5册)
	Conf.VolStart = 2
	Conf.VolEnd = 5

	tests := []struct {
		index    int
		expected bool
		name     string
	}{
		{0, false, "index=0 (vol 1), before start"},
		{1, true, "index=1 (vol 2), at start"},
		{2, true, "index=2 (vol 3), in range"},
		{4, true, "index=4 (vol 5), at end (inclusive)"},
		{5, false, "index=5 (vol 6), after end"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VolumeRange(tt.index); got != tt.expected {
				t.Errorf("VolumeRange(%d) = %v, want %v", tt.index, got, tt.expected)
			}
		})
	}
}

func TestVolumeRange_NegativeEnd(t *testing.T) {
	// 结束册为负数
	Conf.VolStart = 3
	Conf.VolEnd = -1
	// 逻辑: VolEnd < 0 && index > VolStart → index > 3 时返回 false
	// 所以只有 index=2 (vol 3) 有效

	tests := []struct {
		index    int
		expected bool
		name     string
	}{
		{0, false, "index=0, before start"},
		{2, true, "index=2 (vol 3), at start"},
		{3, true, "index=3 (vol 4), still valid (trigger at index>3)"},
		{4, false, "index=4 (vol 5), past negative end"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VolumeRange(tt.index); got != tt.expected {
				t.Errorf("VolumeRange(%d) = %v, want %v", tt.index, got, tt.expected)
			}
		})
	}
}
