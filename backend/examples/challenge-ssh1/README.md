# SSH Challenge

A simple SSH challenge.

Connection example:
`ssh -L 4000:web1:80 test@localhost -p 2222 -o PreferredAuthentications=password -o PubkeyAuthentication=no`
