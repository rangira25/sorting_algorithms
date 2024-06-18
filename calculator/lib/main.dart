import 'package:flutter/material.dart';

void main() => runApp(CalculatorApp());

class CalculatorApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Simple Calculator',
      theme: ThemeData(primarySwatch: Colors.blue),
      home: CalculatorHome(),
    );
  }
}

class CalculatorHome extends StatefulWidget {
  @override
  _CalculatorHomeState createState() => _CalculatorHomeState();
}

class _CalculatorHomeState extends State<CalculatorHome> {
  String _output = "0";
  String _result = "0";
  String? _operation;
  double _num1 = 0;
  double _num2 = 0;

  void _buttonPressed(String buttonText) {
    if (buttonText == "C") {
      _output = "0";
      _num1 = 0;
      _num2 = 0;
      _result = "0";
      _operation = null;
    } else if (buttonText == "+" || buttonText == "-" || buttonText == "*" || buttonText == "/") {
      _num1 = double.parse(_result);
      _operation = buttonText;
      _result = "0";
    } else if (buttonText == ".") {
      if (!_result.contains(".")) {
        _result += buttonText;
      }
    } else if (buttonText == "=") {
      _num2 = double.parse(_result);

      switch (_operation) {
        case "+":
          _result = (_num1 + _num2).toString();
          break;
        case "-":
          _result = (_num1 - _num2).toString();
          break;
        case "*":
          _result = (_num1 * _num2).toString();
          break;
        case "/":
          _result = (_num1 / _num2).toString();
          break;
      }

      _num1 = 0;
      _num2 = 0;
      _operation = null;
    } else {
      _result += buttonText;
    }

    setState(() {
      _output = double.parse(_result).toString();
    });
  }

  Widget _buildButton(String buttonText) {
    return Expanded(
      child: OutlinedButton(
        onPressed: () => _buttonPressed(buttonText),
        child: Text(
          buttonText,
          style: TextStyle(fontSize: 20.0, fontWeight: FontWeight.bold),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Simple Calculator'),
      ),
      body: Column(
        children: <Widget>[
          Container(
            alignment: Alignment.centerRight,
            padding: EdgeInsets.symmetric(vertical: 24.0, horizontal: 12.0),
            child: Text(
              _output,
              style: TextStyle(fontSize: 48.0, fontWeight: FontWeight.bold),
            ),
          ),
          Expanded(
            child: Divider(),
          ),
          Column(
            children: [
              Row(
                children: [
                  _buildButton("7"),
                  _buildButton("8"),
                  _buildButton("9"),
                  _buildButton("/"),
                ],
              ),
              Row(
                children: [
                  _buildButton("4"),
                  _buildButton("5"),
                  _buildButton("6"),
                  _buildButton("*"),
                ],
              ),
              Row(
                children: [
                  _buildButton("1"),
                  _buildButton("2"),
                  _buildButton("3"),
                  _buildButton("-"),
                ],
              ),
              Row(
                children: [
                  _buildButton("."),
                  _buildButton("0"),
                  _buildButton("C"),
                  _buildButton("+"),
                ],
              ),
              Row(
                children: [
                  _buildButton("="),
                ],
              ),
            ],
          ),
        ],
      ),
    );
  }
}
