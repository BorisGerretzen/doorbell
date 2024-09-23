package common

import (
	"crypto/tls"
	"crypto/x509"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"os"
)

type MqttClient struct {
	env    Environment
	Client mqtt.Client
}

func NewMqttHandler(env Environment) *MqttClient {
	return &MqttClient{
		env: env,
	}
}

func (m *MqttClient) Connect() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker("ssl://" + m.env.MqttHost + ":" + m.env.MqttPort)
	opts.SetClientID("doorbell_api")
	opts.SetUsername(m.env.MqttUsername)
	opts.SetPassword(m.env.MqttPassword)

	// Use TLS if provided
	if m.env.MqttCaPath != "" && m.env.MqttCertPath != "" && m.env.MqttKeyPath != "" {
		tlsConfig, err := m.newTlsConfig(m.env.MqttCaPath, m.env.MqttCertPath, m.env.MqttKeyPath)
		if err != nil {
			return err
		}

		opts.SetTLSConfig(tlsConfig)
	}

	m.Client = mqtt.NewClient(opts)
	if token := m.Client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (m *MqttClient) ListenDoorbell(callback func(deviceName string, payload string)) {
	m.Client.Subscribe("doorbell/+", 1, func(client mqtt.Client, msg mqtt.Message) {
		deviceName := msg.Topic()[9:]
		callback(deviceName, string(msg.Payload()))
	})
}

func (m *MqttClient) newTlsConfig(caPath string, certPath string, keyPath string) (*tls.Config, error) {
	certpool := x509.NewCertPool()
	pemCerts, err := os.ReadFile(caPath)
	if err != nil {
		return nil, err
	}
	certpool.AppendCertsFromPEM(pemCerts)

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
	}, nil
}
