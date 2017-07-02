from concurrent import futures
from types import FloatType, StringTypes
import time

import grpc

from ..proto import muni_sign_pb2
from ..proto import muni_sign_pb2_grpc

_ONE_DAY_IN_SECONDS = 60 * 60 * 24

class FakeLCD(object):
    def __init__(self):
        self.text = ""
        self.color = {'red': 0.0, 'green': 0.0, 'blue': 0.0}

    def clear(self):
        self.text = ""

    def set_color(self, red, green, blue):
        assert type(red) is FloatType, 'red color is not a decimal value: %r' % red
        assert type(green) is FloatType, 'green color is not a decimal value: %r' % green
        assert type(blue) is FloatType, 'blue color is not a decimal value: %r' % blue

        assert red >= 0.0 and red <= 1.0, 'red color must be a decimal between 0.0 and 1.0, got %d' % red
        assert green >= 0.0 and green <= 1.0, 'green color must be a decimal between 0.0 and 1.0, got %d' % green 

        assert blue >= 0.0 and blue <= 1.0, 'blue color must be a decimal between 0.0 and 1.0, got %d' % blue 

        self.color = {'red': red, 'green': green, 'blue': blue}
        print 'Set color to %s.' % self.color

    def message(self, msg): 
        assert isinstance(msg, StringTypes), 'message must be a string: %r' % msg

        self.msg = msg
        print 'Set message to \n"""\n%s\n"""' % self.msg

    def __str__(self):
        return self.msg

    def __repr__(self):
        return 'Red: %d, Green: %d, Blue: %d, Message: %s' % (self.color.red, 
                                                              self.color.green, 
                                                              self.color.blue, 
                                                              self.msg)

class DisplayDriver(muni_sign_pb2_grpc.DisplayDriverServicer):

    def __init__(self, lcd):
        self.lcd = lcd

    def Write(self, request, context):
        self.lcd.clear()
        self.lcd.set_color(request.color.red, request.color.green, request.color.blue)
        self.lcd.message(request.message)
        return muni_sign_pb2.Empty()

def serve():
    lcd = FakeLCD()

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    muni_sign_pb2_grpc.add_DisplayDriverServicer_to_server(DisplayDriver(lcd), server)
    server.add_insecure_port('[::]:50051')
    server.start()
    try:
        while True:
            time.sleep(_ONE_DAY_IN_SECONDS)
    except KeyboardInterrupt:
        server.stop(0)

if __name__ == '__main__':
    serve()
