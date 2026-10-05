# Import rclpy, the Python client library that lets this program participate in ROS 2.
import rclpy

# Import Node, the base class that provides ROS 2 node behavior such as subscriptions and logging.
from rclpy.node import Node

# Import the standard Float32 ROS 2 message type.
# The subscriber must use the same message type as the publisher in order to communicate correctly.
from std_msgs.msg import Float32


# Define a ROS 2 node whose job is to listen for wheel-speed messages.
class WheelSpeedMonitor(Node):
    # __init__ runs when we create a WheelSpeedMonitor object.
    def __init__(self):
        # Initialize the parent Node class and give this node the ROS-visible name "wheel_speed_monitor".
        # You can see this name with commands such as `ros2 node list` or in rqt_graph.
        super().__init__("wheel_speed_monitor")

        # Create a ROS 2 subscription and save the returned subscription object in self.sub.
        self.sub = self.create_subscription(
            # This subscriber expects messages of type Float32.
            Float32,

            # This is the exact ROS topic name the subscriber listens to.
            # IMPORTANT: this exercise intentionally uses "/vehicle/wheelspeed" here,
            # while the publisher uses "/vehicle/wheel_speed" with an underscore.
            # ROS topic names must match exactly, so this mismatch prevents communication.
            "/vehicle/wheelspeed",

            # self.on_speed is the callback ROS should run whenever a matching message arrives.
            self.on_speed,

            # 10 is the queue/history depth in the default QoS profile.
            10,
        )

        # Log a startup message so we know the monitor node initialized successfully.
        self.get_logger().info("monitor ready")

    # ROS calls this callback every time a Float32 message arrives on the subscribed topic.
    # msg is the actual Float32 message object ROS delivered to us.
    def on_speed(self, msg):
        # Read msg.data, format it to two decimal places, and print it using the ROS node logger.
        self.get_logger().info(f"received: {msg.data:.2f} m/s")


# main sets up the ROS runtime, creates the node, and keeps it alive.
def main():
    # Initialize rclpy before constructing any ROS nodes.
    rclpy.init()

    # Construct the WheelSpeedMonitor node; its constructor creates the subscription above.
    node = WheelSpeedMonitor()

    # Use try/finally so cleanup happens even when spinning stops because of Ctrl+C or another exit path.
    try:
        # spin enters ROS's event loop.
        # While spinning, rclpy waits for events such as incoming subscription messages and runs callbacks.
        rclpy.spin(node)
    finally:
        # Destroy the node so its ROS resources are released cleanly.
        node.destroy_node()

        # Shut down the ROS 2 Python runtime for this process.
        rclpy.shutdown()


# When this file is executed directly, Python sets __name__ to "__main__".
# If another Python file imports this module, this block does not run automatically.
if __name__ == "__main__":
    # Start the monitor program.
    main()
