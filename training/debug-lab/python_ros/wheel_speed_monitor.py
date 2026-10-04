import rclpy
from rclpy.node import Node
from std_msgs.msg import Float32


class WheelSpeedMonitor(Node):
    def __init__(self):
        super().__init__("wheel_speed_monitor")
        self.sub = self.create_subscription(
            Float32,
            "/vehicle/wheelspeed",
            self.on_speed,
            10,
        )
        self.get_logger().info("monitor ready")

    def on_speed(self, msg):
        self.get_logger().info(f"received: {msg.data:.2f} m/s")


def main():
    rclpy.init()
    node = WheelSpeedMonitor()
    try:
        rclpy.spin(node)
    finally:
        node.destroy_node()
        rclpy.shutdown()


if __name__ == "__main__":
    main()
