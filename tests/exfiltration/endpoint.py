from http.server import HTTPServer, BaseHTTPRequestHandler

class SimpleHTTPRequestHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        content_length = self.headers.get('Content-Length')

        if not content_length:
            self.send_response(400)
            self.end_headers()
            self.wfile.write(b"Missing Content-Length header.")
            return

        content_length = int(content_length)
        filename = self.headers.get('filename', 'uploaded_file')

        # Save the uploaded file to the current directory
        with open(filename, 'wb') as f:
            f.write(self.rfile.read(content_length))

        print(f"Received file: {filename}")

        # Send response
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"File received successfully.")

if __name__ == "__main__":
    port = 8089
    print(f"Starting server on port {port}...")
    httpd = HTTPServer(('0.0.0.0', port), SimpleHTTPRequestHandler)
    httpd.serve_forever()
