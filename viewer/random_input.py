import io
import random
import secrets
import sys
import time

def read_buf(buf):
    i = buf.tell()
    buf.seek(0)
    return buf.read(i)

if __name__ == '__main__':
    sys.stdout.buffer.write(b'info: mount=nvme0n2 total_zone=905 zone_blocks=524288\n')
    sys.stdout.flush()
    buf = io.BytesIO()
    seg_types = [random.randint(0, 5) for i in range(905)]
    while True:
        seg_no = random.randint(0, 926719)
        cur_zone = random.randint(0, 904)
        seg_type = seg_types[cur_zone]
        buf.write(seg_no.to_bytes(4, 'little'))
        buf.write(cur_zone.to_bytes(4, 'little'))
        buf.write(seg_type.to_bytes(4, 'little'))
        buf.write(secrets.token_bytes(64))
        sys.stdout.buffer.write(read_buf(buf))
        sys.stdout.flush()
        buf.seek(0)
#         time.sleep(0.01)
