
# ==========================================
# Stage 1: Build stage (ใช้ image ที่มี Go compiler)
# ==========================================
FROM golang:1.25.0-alpine AS builder

# ติดตั้งเครื่องมือจำเป็นสำหรับคอมไพล์โค้ดบางตัว
RUN apk update && apk add --no-cache git gcc musl-dev

# กำหนด Working Directory ภายใน Container
WORKDIR /app

# คัดลอกไฟล์ dependency ก่อนเพื่อทำ Caching (ช่วยให้ Build ไวขึ้นในครั้งถัดไปถ้า lib ไม่เปลี่ยน)
COPY go.mod go.sum ./
RUN go mod download

# คัดลอกโค้ดทั้งหมดในโปรเจกต์เข้ามา
COPY . .

#  คอมไพล์โค้ดให้กลายเป็นไฟล์ไบนารีเดี่ยวตัวเดียว รองรับ Multi-Architecture (เช่น ARM64 หรือ AMD64) ป้องกันการล็อกสเปกเครื่องปลายทาง
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -ldflags="-w -s" -o cashlog-api ./cmd/api/main.go

# ==========================================
# Stage 2: Final stage (สร้าง Container ตัวจริงที่น้ำหนักเบา)
# ==========================================
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

# สร้าง Non-root User/Group มารองรับการรันแอป (ห้ามใช้สิทธิ์ Root รันบน Production)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# คัดลอกเฉพาะไฟล์ Binary ที่คอมไพล์เสร็จแล้วมาจาก Stage 1
COPY --from=builder /app/cashlog-api .

# สั่งให้ Container สลับไปใช้สิทธิ์ Non-root user ทันที
USER appuser

# เปิดพอร์ตที่ Go Fiber ใช้
EXPOSE 8080

# คำสั่งสำหรับรันแอปพลิเคชัน (เปลี่ยนเป็นรันผ่านโฟลเดอร์ /app)
CMD ["./cashlog-api"]
