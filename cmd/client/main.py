'''
Example client written in python
'''
import socket
import msgpack

def arg_parser(input):
    parts = []
    current = ""
    in_quotes = False

    for char in input:
        if char == '"':
            in_quotes = not in_quotes
            if not in_quotes:
                parts.append(current)
                current = ""
        elif char == ' ' and not in_quotes:
            if current:
                parts.append(current)
                current = ""
        else:
            current += char

    if current:
        parts.append(current)

    if in_quotes:
        raise ValueError("Unmatched quotes in input")

    if not parts:
        raise ValueError("No command entered")

    command = parts[0].upper()
    request = {"command": command}

    if command == "PING":
        if len(parts) > 1:
            raise ValueError("PING does not require any arguments")

    elif command == "ECHO":
        if len(parts) < 2:
            raise ValueError("ECHO requires a message")
        request["message"] = " ".join(parts[1:])

    elif command == "SET":
        if len(parts) < 3:
            raise ValueError("SET requires a key, value, and optional expiry")
        request["key"] = parts[1]
        request["value"] = parts[2]
        if len(parts) > 3:
            try:
                request["exp"] = int(parts[3])
            except ValueError:
                raise ValueError(f"Invalid expiry value: {parts[3]}")

    elif command == "GET":
        if len(parts) < 2:
            raise ValueError("GET requires a key")
        request["key"] = parts[1]

    elif command == "INCR":
        if len(parts) < 3:
            raise ValueError("INCR requires a key and offset")
        request["key"] = parts[1]
        request["offset"] = parts[2]

    else:
        raise ValueError(f"Unknown command: {command}")

    print(request)

    return request

def main():
    try:
        conn = socket.create_connection(("localhost", 6379))
        print("Connected to server. Type commands (e.g., PING, ECHO, SET key value, GET key) and press Enter.")
    except Exception as e:
        print(f"Error connecting to server: {e}")
        return

    try:
        while True:
            user_input = input(">> ").strip()
            if not user_input:
                continue

            try:
                request = arg_parser(user_input)
            except ValueError as e:
                print(f"Error: {e}")
                continue

            try:
                data = msgpack.packb(request)
                conn.sendall(data)

                response = conn.recv(4096)
                server_response = msgpack.unpackb(response, strict_map_key=False)
                print(server_response)
                status = server_response.get("status")
                if status == "OK":
                    message = server_response.get("message")
                    value = server_response.get("value")
                    if message:
                        print("Server:", message)
                    elif value:
                        print("Server:", value)
                    else:
                        print("Server: OK")
                elif status == "ERROR":
                    print("Server Error:", server_response.get("message"))
                else:
                    print("Unexpected server response:", server_response)

            except Exception as e:
                print(f"Error communicating with server: {e}")

    except KeyboardInterrupt:
        print("\nClosing connection.")
    finally:
        conn.close()

if __name__ == "__main__":
    main()
