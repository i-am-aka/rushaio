#!/usr/bin/env python3

import frida
import json
import time
import sys


from queue import Queue  # or queue in Python 3
import threading

class PrintThread(threading.Thread):
    def __init__(self, queue, fn):
        threading.Thread.__init__(self)
        self.queue = queue
        self.fn = fn
        # self.file = open(fn, 'w')

    def run(self):
      while True:
        result = self.queue.get()
        with open(self.fn, 'a+') as f:
          f.write(result + '\n')
        # print(result)
        self.queue.task_done()

trace_file_queue = Queue()
t = PrintThread(trace_file_queue, f'trace-{int(time.time())}.txt')
t.setDaemon(True)
t.start()


device = frida.get_usb_device()
pid = device.spawn(["com.WinnersCircle"])
device.resume(pid)
# time.sleep(1) #Without it Java.perform silently fails
session = device.attach(pid)

def on_message(message, data):
  # if message['type'] == 'send':
  #     msg = "[*] {0}".format(message['payload'])
  # else:
  #     msg = message

  trace_file_queue.put_nowait(str(message['payload']))
  # f.write((msg + '\n').encode())
  # f.flush()
  # n += 1


script = session.create_script(open("frida-trace2.js").read())
script.on('message', on_message)
script.load()

sys.stdin.read()
