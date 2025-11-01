package naqel

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"prayog-serviceability-service/internal/shared/config"

	"github.com/sirupsen/logrus"
)

// NaqelClient handles SOAP communication with Naqel API
type NaqelClient struct {
	httpClient *http.Client
	config     config.NaqelConfig
	logger     *logrus.Logger
}

// NewNaqelClient creates a new Naqel SOAP client
func NewNaqelClient(config config.NaqelConfig) *NaqelClient {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	logger.WithFields(logrus.Fields{
		"partner":      "Naqel",
		"base_url":     config.BaseURL,
		"client_id":    redactField(config.ClientID),
		"load_type_id": config.LoadTypeID,
		"timeout":      config.Timeout,
	}).Info("Creating Naqel SOAP client")

	return &NaqelClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}

// redactField safely redacts sensitive fields for logging
func redactField(field string) string {
	if field == "" {
		return "[EMPTY]"
	}
	if len(field) <= 3 {
		return "[REDACTED]"
	}
	return field[:3] + "***"
}

// GetTransitDays calls Naqel SOAP API to get transit days between two city codes
func (c *NaqelClient) GetTransitDays(ctx context.Context, originCityCode, destinationCityCode, originStationCode, destinationStationCode string) (int, error) {
	// Build SOAP request
	soapRequest := c.buildSOAPRequest(originCityCode, destinationCityCode, originStationCode, destinationStationCode)
	c.logger.Info("soapRequest", soapRequest)
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.BaseURL, bytes.NewBufferString(soapRequest))
	if err != nil {
		return 0, fmt.Errorf("failed to create SOAP request: %w", err)
	}

	// Set SOAP headers
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", `"http://tempuri.org/GetTransitDays"`)

	c.logger.WithFields(logrus.Fields{
		"partner":            "Naqel",
		"origin_city_code":   originCityCode,
		"destination_city_code": destinationCityCode,
		"url":                c.config.BaseURL,
	}).Info("Making Naqel SOAP API request")

	// Execute request with retry logic
	var resp *http.Response
	var lastErr error
	maxRetries := 1
	

	// for attempt := 0; attempt <= maxRetries; attempt++ {
	// 	if attempt > 0 {
	// 		delay := c.config.RetryDelay * time.Duration(attempt)
	// 		time.Sleep(delay)
	// 	}

		resp, lastErr = c.httpClient.Do(req)
	// 	if lastErr == nil && resp.StatusCode < 500 {
	// 		break // Success or client error (don't retry)
	// 	}

	// 	if resp != nil {
	// 		resp.Body.Close()
	// 	}
	// }

	if lastErr != nil {
		return 0, fmt.Errorf("SOAP request failed after %d retries: %w", maxRetries, lastErr)
	}

	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read SOAP response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"partner":     "Naqel",
		"status_code": resp.StatusCode,
		"body_size":   len(body),
	}).Info("Received Naqel SOAP API response")

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("SOAP API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse SOAP XML response
	var soapResponse GetTransitDaysResponse
	if err := xml.Unmarshal(body, &soapResponse); err != nil {
		c.logger.WithFields(logrus.Fields{
			"partner":  "Naqel",
			"response": string(body),
			"error":    err.Error(),
		}).Warn("Failed to parse SOAP response")
		return 0, fmt.Errorf("failed to parse SOAP response: %w", err)
	}

	if soapResponse.Days <= 0 {
		return 0, fmt.Errorf("invalid transit days returned: %d", soapResponse.Days)
	}

	c.logger.WithFields(logrus.Fields{
		"partner":            "Naqel",
		"origin_city_code":   originCityCode,
		"destination_city_code": destinationCityCode,
		"transit_days":       soapResponse.Days,
	}).Info("Successfully retrieved transit days from Naqel")

	return soapResponse.Days, nil
}

// buildSOAPRequest builds the SOAP XML envelope for GetTransitDays
func (c *NaqelClient) buildSOAPRequest(originCityCode, destinationCityCode, originStationCode, destinationStationCode string) string {
	// Default client address values (can be configured if needed)
	clientAddress := ClientAddress{
		PhoneNumber:     "",
		NationalAddress: "",
		POBox:           "",
		ZipCode:         "",
		Fax:             "",
		Latitude:        "",
		Longitude:       "",
		ShipperName:     "Test Shipper",
		FirstAddress:    "Test Address",
		Location:        "",
		CountryCode:     "",
		CityCode:        originCityCode,
	}

	clientContact := ClientContact{
		Name:        "Test Client",
		Email:       "test@example.com",
		PhoneNumber: "0500000000",
		MobileNo:    "0500000000",
	}

	clientInfo := ClientInfo{
		ClientAddress: clientAddress,
		ClientContact: clientContact,
		ClientID:      c.config.ClientID,
		Password:      c.config.Password,
		Version:       "9.0",
	}

	// Build XML manually to match SOAP format
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="utf-8"?>`)
	sb.WriteString(`<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"`)
	sb.WriteString(` xmlns:xsd="http://www.w3.org/2001/XMLSchema"`)
	sb.WriteString(` xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">`)
	sb.WriteString(`<soap:Body>`)
	sb.WriteString(`<GetTransitDays xmlns="http://tempuri.org/">`)
	sb.WriteString(`<ClientInfo>`)
	sb.WriteString(`<ClientAddress>`)
	sb.WriteString(fmt.Sprintf(`<PhoneNumber>%s</PhoneNumber>`, escapeXML(clientInfo.ClientAddress.PhoneNumber)))
	sb.WriteString(fmt.Sprintf(`<NationalAddress>%s</NationalAddress>`, escapeXML(clientInfo.ClientAddress.NationalAddress)))
	sb.WriteString(fmt.Sprintf(`<POBox>%s</POBox>`, escapeXML(clientInfo.ClientAddress.POBox)))
	sb.WriteString(fmt.Sprintf(`<ZipCode>%s</ZipCode>`, escapeXML(clientInfo.ClientAddress.ZipCode)))
	sb.WriteString(fmt.Sprintf(`<Fax>%s</Fax>`, escapeXML(clientInfo.ClientAddress.Fax)))
	sb.WriteString(fmt.Sprintf(`<Latitude>%s</Latitude>`, escapeXML(clientInfo.ClientAddress.Latitude)))
	sb.WriteString(fmt.Sprintf(`<Longitude>%s</Longitude>`, escapeXML(clientInfo.ClientAddress.Longitude)))
	sb.WriteString(fmt.Sprintf(`<ShipperName>%s</ShipperName>`, escapeXML(clientInfo.ClientAddress.ShipperName)))
	sb.WriteString(fmt.Sprintf(`<FirstAddress>%s</FirstAddress>`, escapeXML(clientInfo.ClientAddress.FirstAddress)))
	sb.WriteString(fmt.Sprintf(`<Location>%s</Location>`, escapeXML(clientInfo.ClientAddress.Location)))
	sb.WriteString(fmt.Sprintf(`<CountryCode>%s</CountryCode>`, escapeXML(clientInfo.ClientAddress.CountryCode)))
	sb.WriteString(fmt.Sprintf(`<CityCode>%s</CityCode>`, escapeXML(clientInfo.ClientAddress.CityCode)))
	sb.WriteString(`</ClientAddress>`)
	sb.WriteString(`<ClientContact>`)
	sb.WriteString(fmt.Sprintf(`<Name>%s</Name>`, escapeXML(clientInfo.ClientContact.Name)))
	sb.WriteString(fmt.Sprintf(`<Email>%s</Email>`, escapeXML(clientInfo.ClientContact.Email)))
	sb.WriteString(fmt.Sprintf(`<PhoneNumber>%s</PhoneNumber>`, escapeXML(clientInfo.ClientContact.PhoneNumber)))
	sb.WriteString(fmt.Sprintf(`<MobileNo>%s</MobileNo>`, escapeXML(clientInfo.ClientContact.MobileNo)))
	sb.WriteString(`</ClientContact>`)
	sb.WriteString(fmt.Sprintf(`<ClientID>%s</ClientID>`, escapeXML(clientInfo.ClientID)))
	sb.WriteString(fmt.Sprintf(`<Password>%s</Password>`, escapeXML(clientInfo.Password)))
	sb.WriteString(fmt.Sprintf(`<Version>%s</Version>`, escapeXML(clientInfo.Version)))
	sb.WriteString(`</ClientInfo>`)
	sb.WriteString(fmt.Sprintf(`<Origin>%s</Origin>`, escapeXML(originStationCode)))
	sb.WriteString(fmt.Sprintf(`<Destination>%s</Destination>`, escapeXML(destinationStationCode)))
	sb.WriteString(fmt.Sprintf(`<loadtypeID>%d</loadtypeID>`, c.config.LoadTypeID))
	sb.WriteString(`</GetTransitDays>`)
	sb.WriteString(`</soap:Body>`)
	sb.WriteString(`</soap:Envelope>`)

	return sb.String()
}

// escapeXML escapes special XML characters
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

