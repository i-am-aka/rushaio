import zmq
import threading
from zmq.utils.monitor import recv_monitor_message

# EVENT_MAP = {}
# print("Event names:")
# for name in dir(zmq):
#   if name.startswith('EVENT_'):
#     value = getattr(zmq, name)
#     print("%21s : %4i" % (name, value))
#     EVENT_MAP[value] = name

PUB_PORT = "42423"
SUB_TAIL_PORT = "42425"
context = zmq.Context()

def get_task_pub_socket():
  socket = context.socket(zmq.PUB)
  socket.bind("tcp://127.0.0.1:%s" % PUB_PORT)
  evt = threading.Event()
  def monitor(m,evt):
    while m.poll():
      mevt = recv_monitor_message(m)
      # print("Event: {}".format(mevt))
      if mevt['event'] == 4096:
        evt.set()
        break

  t = threading.Thread(target=monitor, args=(socket.get_monitor_socket(),evt))
  t.start()
  return socket, evt

def get_task_tail_socket():
  socket = context.socket(zmq.SUB)
  socket.connect("tcp://127.0.0.1:%s" % SUB_TAIL_PORT)
  socket.setsockopt(zmq.SUBSCRIBE, b"")
  poller = zmq.Poller()
  poller.register(socket, zmq.POLLIN)
  return poller
