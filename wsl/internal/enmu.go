package internal

import "fmt"

// Code just copy from https://github.com/ubuntu/GoWSL

const (
	Stopped = iota
	Running
	Installing
	Uninstalling
	NotRegistered
	Error
)

func String2Code(stat_str string) (int, error) {
	switch stat_str {
	case "Stopped":
		return Stopped, nil
	case "Running":
		return Running, nil
	case "Installing":
		return Installing, nil
	case "Uninstalling":
		return Uninstalling, nil
	case "NotRegistered":
		return NotRegistered, nil
	}
	return Error, fmt.Errorf("could not parse %s to code", stat_str)
}

func Code2State(code int) string {
	switch code {
	case Stopped:
		return "Stopped"
	case Running:
		return "Running"
	case Installing:
		return "Installing"
	case Uninstalling:
		return "Uninstalling"
	case NotRegistered:
		return "NotRegistered"
	}
	return fmt.Sprintf("code %d not implemented", code)
}
