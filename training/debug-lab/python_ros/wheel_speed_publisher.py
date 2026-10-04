import math
import rclpy
from rclpy.node import Node
from std_msgs.msg import Float32

class WheelSpeedPublisher(Node):
    def __init__(self):
        super().__init__("wheel_speed_sensor")
        self.pub = self.create_publisher(Float32, "/vehicle/wheel_speed", 10)
        self.t = 0.0
        self.timer = self.create_timer(0.25, self.tick)

    def tick(self):
        msg = Float32()
        msg.data = 12.0 + 2.0 * math.sin(self.t)
        self.pub.publish(msg)
        self.get_logger().info(f"published: {msg.data:.2f} m/s")
        self.t += 0.25

def main():
    rclpy.init()
    node = WheelSpeedPublisher()
    try:
        rclpy.spin(node)
    finally:
        node.destroy_node()
        rclpy.shutdown()

if __name__ == "__main__":
    main()
