package node

import "testing"

// 内存自适应限流的判定必须是确定性的：连续 3 次高位才降，连续 3 次低位才恢复。
// 这里把阈值行为钉住——改 memHighWater/memSamplesHigh 会直接让测试失败。
func TestDecideThrottleAndRestore(t *testing.T) {
	throttled := false
	hs, ls := 0, 0
	// 两次高位还不够（避免抖动）
	for i := 0; i < 2; i++ {
		var changed bool
		throttled, hs, ls, changed = decide(95, 3, throttled, hs, ls)
		if changed || throttled {
			t.Fatalf("sample %d: throttled too early (hs=%d)", i+1, hs)
		}
	}
	// 第三次才降
	var changed bool
	throttled, hs, ls, changed = decide(95, 3, throttled, hs, ls)
	if !throttled || !changed {
		t.Fatalf("expected throttle after 3 high samples, got throttled=%v changed=%v", throttled, changed)
	}
	// 中间值不恢复也不加重
	throttled, hs, ls, changed = decide(80, 3, throttled, hs, ls)
	if !throttled || changed {
		t.Fatalf("mid value must keep state, got throttled=%v changed=%v", throttled, changed)
	}
	// 连续三次低位才恢复
	for i := 0; i < 2; i++ {
		throttled, hs, ls, changed = decide(60, 3, throttled, hs, ls)
		if !throttled {
			t.Fatalf("restored too early at sample %d", i+1)
		}
	}
	throttled, _, _, changed = decide(60, 3, throttled, hs, ls)
	if throttled || !changed {
		t.Fatalf("expected restore after 3 low samples, got throttled=%v changed=%v", throttled, changed)
	}
}
