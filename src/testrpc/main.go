package main

import (
    "crypto/rand"
    "crypto/rsa"
    "fmt"
    "crypto/x509"
    "encoding/pem"
)


func main() {
    // crypto/rand.Reader is a good source of entropy for blinding the RSA
    // operation.
    rng := rand.Reader

    key, err := rsa.GenerateKey(rng, 512)
    if err != nil {
        panic(err)
    }

    pemdata := pem.EncodeToMemory(
        &pem.Block{
            Type: "RSA PRIVATE KEY",
            Bytes: x509.MarshalPKCS1PrivateKey(key),
        },
    )

    f := string(pemdata)
    fmt.Println(f)

    block, _ := pem.Decode([]byte(f))
    println(block.Type)
    pr, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
    println(pr)



    //
    //message := []byte("message to be signed")
    //
    //// Only small messages can be signed directly; thus the hash of a
    //// message, rather than the message itself, is signed. This requires
    //// that the hash function be collision resistant. SHA-256 is the
    //// least-strong hash function that should be used for this at the time
    //// of writing (2016).
    //hashed := sha256.Sum256(message)
    //
    //signature, err := rsa.SignPKCS1v15(rng, rsaPrivateKey, crypto.SHA256, hashed[:])
    //if err != nil {
    //    fmt.Fprintf(os.Stderr, "Error from signing: %s\n", err)
    //    return
    //}
    //
    //fmt.Printf("Signature: %x\n", signature)



    //client, err := rpc.DialHTTP("tcp", "localhost:7654")
    //if err != nil {
    //    log.Fatal("dialing:", err)
    //}
    //
    //if len(os.Args) < 2 {
    //    fmt.Printf("Usage: %s command [args...]\n", os.Args[0])
    //    return
    //}
    //
    //aCmd := os.Args[1]
    //aArgs := []string{}
    //if len(os.Args) > 2 {
    //    aArgs = os.Args[2:]
    //}
    //
    //args := &srpc.ExecArgs{
    //    Cmd: aCmd,
    //    Args: aArgs,
    //}
    //reply := srpc.ExecReply{}
    //err = client.Call("SonarRpc.Exec", args, &reply)
    //if err != nil {
    //    log.Fatal("arith error:", err)
    //}
    //
    //fmt.Printf("Out: %s\n", reply.Stdout)
    //fmt.Printf("Err: %s\n", reply.Stderr)
}
