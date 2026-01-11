package ssh

import (
    "golang.org/x/crypto/ssh"
    "os"
    "time"
)

func ExecuteCommand(ip, user, privateKeyPath, command string) (string, error) {
    key, err := os.ReadFile(privateKeyPath)
    if err != nil {
        return "", err
    }

    signer, err := ssh.ParsePrivateKey(key)
    if err != nil {
        return "", err
    }

    config := &ssh.ClientConfig{
        User: user,
        Auth: []ssh.AuthMethod{
            ssh.PublicKeys(signer),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For Milestone 1 only
        Timeout:         10 * time.Second,
    }

    client, err := ssh.Dial("tcp", ip+":22", config)
    if err != nil {
        return "", err
    }
    defer client.Close()

    session, err := client.NewSession()
    if err != nil {
        return "", err
    }
    defer session.Close()

    output, err := session.CombinedOutput(command)
    return string(output), err
}