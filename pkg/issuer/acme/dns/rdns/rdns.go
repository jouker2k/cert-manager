package rdns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type DNSProvider struct {
	client DnsClient
}

func NewDNSProvider(apiEndpoint, token string) (*DNSProvider, error) {
	dnsClient := DnsClient{
		httpClient: http.DefaultClient,
		base:       apiEndpoint,
		token:      token,
	}
	return &DNSProvider{
		client: dnsClient,
	}, nil
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	return d.client.SetTXTRecord(domain, keyAuth)
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	return d.client.DeleteDNSRecord(domain)
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 180 * time.Second, 5 * time.Second
}

type DnsClient struct {
	httpClient *http.Client
	base       string
	token      string
}

func (d *DnsClient) SetTXTRecord(domain, text string) error {
	url := fmt.Sprintf("%s/domain/_acme-challenge.%s/txt", d.base, domain)
	payload := map[string]string{
		"text": text,
	}
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.token))
	request.Header.Set("Content-Type", "application/json")
	resp, err := d.httpClient.Do(request)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("expect 200, got %v. Error: %s", resp.StatusCode, string(data))
	}
	return nil
}

func (d *DnsClient) DeleteDNSRecord(domain string) error {
	url := fmt.Sprintf("%s/domain/_acme-challenge.%s/txt", d.base, domain)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.token))
	request.Header.Set("Content-Type", "application/json")
	_, err = d.httpClient.Do(request)
	return err
}
