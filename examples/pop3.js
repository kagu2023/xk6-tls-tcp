import tcpclient from 'k6/x/tcpclient';
import { check } from 'k6';

export default function () {
    let conn = tcpclient.connect(true, '<hostname>:995', false);
    verify('banner', conn.readLine(), '+OK');

    sendThenVerify(conn, 'USER', '<username>');
    sendThenVerify(conn, 'PASS', '<password>');
    sendThenVerify(conn, 'STAT');
    sendThenVerify(conn, 'QUIT');

    conn.close();
}

function verify(command, response, expected) {
    check(response, {
        [`verify ${command}`]: (response) => response.includes(expected)
    });
}

function sendThenVerify(conn, command, payload = '', expected = '+OK') {
    let fullCommand = command;
    if (payload) {
        fullCommand += ' ' + payload;
    }

    conn.writeStringCRLine(fullCommand);
    verify(command, conn.readLine(), expected);
}