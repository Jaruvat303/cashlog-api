package pkg

// สร้าง helper function สำหรับจัดการแปลง pointer
func PTR[T any](v T) *T { return &v }
