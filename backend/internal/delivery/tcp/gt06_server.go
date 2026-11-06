package tcp

import (
	"encoding/binary"
	"encoding/hex"
	"log"
	"net"
	"time"

	"example.com/hris/backend/internal/usecase"
)

type GT06Server struct {
	port    string
	usecase usecase.GpsUsecase
}

func NewGT06Server(port string, uc usecase.GpsUsecase) *GT06Server {
	return &GT06Server{
		port:    port,
		usecase: uc,
	}
}

func (s *GT06Server) Start() error {
	addr := ":" + s.port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("GT06 TCP Server listening on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting TCP connection: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *GT06Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)

	// Set initial read timeout
	conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

	var currentIMEI string

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			break
		}

		// Reset read deadline on successful read
		conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

		data := buffer[:n]

		// Find start bits 0x78 0x78 or 0x79 0x79
		if n < 10 {
			continue // too short
		}

		if (data[0] == 0x78 && data[1] == 0x78) || (data[0] == 0x79 && data[1] == 0x79) {
			protocolNum := data[3]

			switch protocolNum {
			case 0x01: // Login Message
				imeiBytes := data[4:12]
				currentIMEI = hex.EncodeToString(imeiBytes)
				// Remove leading zero if present
				if currentIMEI[0] == '0' {
					currentIMEI = currentIMEI[1:]
				}
				log.Printf("GPS Device Logged In: %s", currentIMEI)

				// Send Response ACK
				s.sendResponse(conn, protocolNum, data[n-6:n-4]) // Serial number is at n-6, n-5

			case 0x12: // Location Data
				s.parseLocation(currentIMEI, data)

			case 0x13: // Heartbeat
				// Send Response ACK
				s.sendResponse(conn, protocolNum, data[n-6:n-4])
			}
		}
	}
}

func (s *GT06Server) parseLocation(imei string, data []byte) {
	if len(data) < 22 || imei == "" {
		return
	}

	// Parse Datetime
	year := int(data[4]) + 2000
	month := time.Month(data[5])
	day := int(data[6])
	hour := int(data[7])
	min := int(data[8])
	sec := int(data[9])

	timestamp := time.Date(year, month, day, hour, min, sec, 0, time.UTC)

	// Lengths and conversions are according to GT06 protocol
	latBytes := binary.BigEndian.Uint32(data[11:15])
	lngBytes := binary.BigEndian.Uint32(data[15:19])

	speed := float64(data[19])
	courseStatus := binary.BigEndian.Uint16(data[20:22])

	course := float64(courseStatus & 0x3FF)
	isAccOn := (courseStatus & 0x4000) != 0

	// Convert raw lat/lng to decimal degrees
	latitude := float64(latBytes) / 30000.0 / 60.0
	longitude := float64(lngBytes) / 30000.0 / 60.0

	// Determine hemisphere
	if (courseStatus & 0x0400) == 0 { // South
		latitude = -latitude
	}
	if (courseStatus & 0x0800) == 0 { // West
		longitude = -longitude
	}

	// Save to DB via usecase
	err := s.usecase.ProcessLocationData(imei, latitude, longitude, speed, course, isAccOn, timestamp)
	if err != nil {
		log.Printf("Error saving GPS location for %s: %v", imei, err)
	}
}

func (s *GT06Server) sendResponse(conn net.Conn, protocolNum byte, serialNum []byte) {
	response := []byte{0x78, 0x78, 0x05, protocolNum, serialNum[0], serialNum[1]}
	crc := crc16(response[2:]) // calculate CRC from length to serialNum

	response = append(response, byte(crc>>8), byte(crc&0xFF), 0x0D, 0x0A)
	conn.Write(response)
}

func crc16(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
