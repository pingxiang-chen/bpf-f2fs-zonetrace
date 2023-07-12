import io
import random
import secrets
import sys
import time

if __name__ == '__main__':
    sys.stdout.buffer.write(b'info: total_zone=905 zone_blocks=524288\n')
    buf = io.BytesIO()
    while True:
        seg_no = random.randint(0, 1023)
        cur_zone = random.randint(0, 904)
        seg_type = random.randint(0, 5)
        buf.write(f"update_sit_entry segno: {seg_no} cur_zone:{cur_zone} seg_type:{seg_type}\n".encode())
        buf.write(secrets.token_bytes(64))
        buf.write(b'\n')
        i = buf.tell()
        buf.seek(0)
        sys.stdout.buffer.write(buf.read(i))
        buf.seek(0)
