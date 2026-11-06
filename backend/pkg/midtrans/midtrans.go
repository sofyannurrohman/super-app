package midtrans

import (
	"fmt"
	"os"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
)

type MidtransClient struct {
	CoreAPI coreapi.Client
}

func NewMidtransClient() *MidtransClient {
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	env := midtrans.Sandbox
	if os.Getenv("MIDTRANS_IS_PRODUCTION") == "true" {
		env = midtrans.Production
	}

	client := coreapi.Client{}
	client.New(serverKey, env)

	return &MidtransClient{
		CoreAPI: client,
	}
}

// ExtractQRISData extracts the QR Code Image URL from a Midtrans charge response.
//
// For Sandbox: Midtrans returns the "generate-qr-code" image URL.
//   - This URL is what the Midtrans simulator (https://simulator.sandbox.midtrans.com/v2/qris/index) requires.
//   - The mobile app renders this as Image.network (detects it starts with "http").
//
// For Production: Midtrans returns QRString (raw EMVCo string).
//   - The mobile app renders this as QrImageView (a scannable QR code image).
//
// Priority: generate-qr-code-v2 (Sandbox simulator-compatible) → generate-qr-code → QRString (Production raw EMVCo) → qr-string action
func ExtractQRISData(resp *coreapi.ChargeResponse) string {
	if resp == nil {
		return ""
	}

	// 1. Check if there is a "generate-qr-code-v2" action URL (Sandbox GoPay/QRIS Simulator compatible URL)
	for _, action := range resp.Actions {
		if action.Name == "generate-qr-code-v2" {
			return action.URL
		}
	}

	// 2. Check if there is a "generate-qr-code" action URL
	for _, action := range resp.Actions {
		if action.Name == "generate-qr-code" {
			return action.URL
		}
	}

	// 3. Fallback to raw EMVCo QR string (always present in Production)
	if resp.QRString != "" {
		return resp.QRString
	}

	// 4. Fallback: qr-string action
	for _, action := range resp.Actions {
		if action.Name == "qr-string" {
			return action.URL
		}
	}

	return ""
}

func (m *MidtransClient) CreateQRIS(orderID string, amount int64) (*coreapi.ChargeResponse, error) {
	req := &coreapi.ChargeReq{
		PaymentType: coreapi.PaymentTypeQris,
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: amount,
		},
		CustomerDetails: &midtrans.CustomerDetails{
			FName: "Customer",
			LName: orderID,
			Email: "customer@example.com",
		},
	}

	resp, err := m.CoreAPI.ChargeTransaction(req)
	if err != nil {
		return resp, err
	}

	fmt.Printf("DEBUG MIDTRANS QRIS Response — TransactionID: %s | QRString: %s | Actions: %+v\n",
		resp.TransactionID, resp.QRString, resp.Actions)

	return resp, nil
}

func (m *MidtransClient) CreateBankTransfer(orderID string, amount int64, bank string) (*coreapi.ChargeResponse, error) {
	fmt.Printf("DEBUG MIDTRANS: Creating Bank Transfer for OrderID: %s, Amount: %d, Bank: %s\n", orderID, amount, bank)

	if bank == "mandiri" {
		req := &coreapi.ChargeReq{
			PaymentType: coreapi.PaymentTypeEChannel,
			TransactionDetails: midtrans.TransactionDetails{
				OrderID:  orderID,
				GrossAmt: amount,
			},
			EChannel: &coreapi.EChannelDetail{
				BillInfo1: "Pembayaran Sales",
				BillInfo2: "Perusahaan Demo",
			},
			CustomerDetails: &midtrans.CustomerDetails{
				FName: "Customer",
				LName: orderID,
				Email: "customer@example.com",
			},
		}
		resp, err := m.CoreAPI.ChargeTransaction(req)
		if err != nil {
			fmt.Printf("ERROR MIDTRANS Mandiri: %v\n", err)
			return resp, err
		}
		return resp, nil
	}

	if bank == "permata" {
		req := &coreapi.ChargeReq{
			PaymentType: coreapi.PaymentTypeBankTransfer,
			TransactionDetails: midtrans.TransactionDetails{
				OrderID:  orderID,
				GrossAmt: amount,
			},
			BankTransfer: &coreapi.BankTransferDetails{
				Bank: midtrans.BankPermata,
			},
			CustomerDetails: &midtrans.CustomerDetails{
				FName: "Customer",
				LName: orderID,
				Email: "customer@example.com",
			},
		}
		resp, err := m.CoreAPI.ChargeTransaction(req)
		if err != nil {
			return resp, err
		}
		return resp, nil
	}

	var bankEnum midtrans.Bank
	switch bank {
	case "bca":
		bankEnum = midtrans.BankBca
	case "bni":
		bankEnum = midtrans.BankBni
	case "bri":
		bankEnum = midtrans.BankBri
	default:
		bankEnum = midtrans.BankBca
	}

	req := &coreapi.ChargeReq{
		PaymentType: coreapi.PaymentTypeBankTransfer,
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: amount,
		},
		BankTransfer: &coreapi.BankTransferDetails{
			Bank: bankEnum,
		},
		CustomerDetails: &midtrans.CustomerDetails{
			FName: "Customer",
			LName: orderID,
			Email: "customer@example.com",
		},
	}

	resp, err := m.CoreAPI.ChargeTransaction(req)
	if err != nil {
		fmt.Printf("ERROR MIDTRANS %s: %v\n", bank, err)
		return resp, err
	}
	return resp, nil
}

func (m *MidtransClient) CheckTransaction(orderID string) (*coreapi.TransactionStatusResponse, error) {
	resp, midErr := m.CoreAPI.CheckTransaction(orderID)
	if midErr != nil {
		return nil, fmt.Errorf("midtrans error: %s (code: %d)", midErr.Message, midErr.StatusCode)
	}
	return resp, nil
}
