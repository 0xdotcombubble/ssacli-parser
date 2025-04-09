# Telegraf Disk Parser

A command-line utility for parsing disk information and outputting in Telegraf-compatible format (InfluxDB Line Protocol).

## Features

- Parses disk metrics including usage, temperature, and status
- Outputs in InfluxDB Line Protocol format for easy integration with Telegraf
- Supports reading from files or standard input
- Tracks metrics by box and bay positions

## Metrics Collected

### Tags
- `box`: Box identifier (integer)
- `bay`: Bay identifier (integer)
- `size_gb`: Disk size in GB (integer)

### Fields
- `usage_remaining`: Remaining disk usage percentage
- `estimated_life_remaining`: Estimated remaining life in days
- `status`: Disk status (1.0 = OK, 0.0 = Not OK)
- `current_temperature`: Current disk temperature
- `maximum_temperature`: Maximum recorded temperature
- `power_on_hours`: Hours the disk has been powered on

## Getting Started

### Prerequisites

- Go 1.24 or higher

### Installation

1. Clone the repository
   ```
   git clone https://github.com/0xdotcombubble/ssacli_telgraf.git
   ```

2. Navigate to the project directory
   ```
   cd ssacli_telgraf
   ```

3. Build the project
   ```
   go build
   ```

## Contributing

If you want to contribute to this project:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Before pushing changes, make sure to sync with the remote repository:
```
git pull origin main
```

## Usage

### Reading from a file
```
./telegraf_diskparser -file input.txt
```

### Reading from standard input
```
cat input.txt | ./telegraf_diskparser
```

## Example Output

```
disk,box=1,bay=1,size_gb=1024 current_temperature=35.000000,estimated_life_remaining=730.000000,maximum_temperature=38.000000,power_on_hours=8760.000000,status=1.000000,usage_remaining=94.500000
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
