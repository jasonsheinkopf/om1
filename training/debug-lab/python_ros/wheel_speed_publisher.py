# Import Python's math module so we can use math.sin() to generate a changing fake wheel speed.
import math

# Import rclpy, the standard Python client library for ROS 2.
# It provides ROS initialization, spinning, shutdown, and other runtime functions.
import rclpy

# Import Node, the base class used to create ROS 2 nodes in Python.
from rclpy.node import Node

# Import the standard ROS 2 Float32 message type.
# Each published message will contain one 32-bit floating-point value in its `.data` field.
from std_msgs.msg import Float32


# Define a new ROS 2 node class named WheelSpeedPublisher.
# Inheriting from Node gives this class ROS capabilities such as publishers, timers, and logging.
class WheelSpeedPublisher(Node):
    # __init__ is Python's constructor; it runs when we create a WheelSpeedPublisher object.
    def __init__(self):
        # Initialize the parent Node class and register this node with the name "wheel_speed_sensor".
        # This node name is what tools such as `ros2 node list` and rqt_graph will display.
        super().__init__("wheel_speed_sensor")

        # Create a ROS 2 publisher.
        # Float32 says which message type this publisher sends.
        # "/vehicle/wheel_speed" is the topic name.
        # 10 is the queue/history depth used by the default QoS profile.
        self.pub = self.create_publisher(Float32, "/vehicle/wheel_speed", 10)

        # Store a simple time-like variable used as the input to the sine wave.
        # We begin at zero radians.
        self.t = 0.0

        # Create a timer that calls self.tick every 0.25 seconds.
        # That means this publisher attempts to send data four times per second (4 Hz).
        self.timer = self.create_timer(0.25, self.tick)

    # tick is the callback that ROS invokes whenever the 0.25-second timer fires.
    def tick(self):
        # Construct a new empty Float32 ROS message object.
        msg = Float32()

        # Put a simulated speed value into the message's data field.
        # The signal is centered at 12.0 m/s and oscillates by plus/minus 2.0 m/s.
        msg.data = 12.0 + 2.0 * math.sin(self.t)

        # Publish the message on the topic configured in self.pub.
        # At this point ROS 2/DDS handles delivering it to any compatible subscribers.
        self.pub.publish(msg)

        # Write a human-readable log line so we can verify that this node is actively publishing.
        self.get_logger().info(f"published: {msg.data:.2f} m/s")

        # Advance our fake time variable by 0.25 for the next sine-wave sample.
        # This matches the timer period numerically, so the wave progresses roughly in real-time seconds.
        self.t += 0.25


# main contains the setup and event-loop code for running the node as a standalone Python program.
def main():
    # Initialize the ROS 2 Python runtime for this process.
    # This must happen before creating ROS nodes.
    rclpy.init()

    # Construct our publisher node; its __init__ method creates the publisher and timer.
    node = WheelSpeedPublisher()

    # Use try/finally so ROS cleanup still happens when spinning stops because of Ctrl+C or another exit path.
    try:
        # spin keeps this process alive and lets ROS execute callbacks such as self.tick.
        # Without spin, the script would create the node and then immediately exit.
        rclpy.spin(node)
    finally:
        # Explicitly destroy the node and release its ROS resources.
        node.destroy_node()

        # Shut down the rclpy runtime for this process.
        rclpy.shutdown()


# Python sets __name__ to "__main__" when this file is executed directly.
# This condition prevents main() from running automatically if the file is merely imported as a module.
if __name__ == "__main__":
    # Start the program.
    main()
