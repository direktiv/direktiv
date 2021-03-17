import 'package:flutter/widgets.dart';
import 'package:flutter/material.dart';

class Workflow extends StatelessWidget {
  Workflow({@required this.workflow});
  final String workflow;

  @override
  Widget build(BuildContext context) {
    return Container(
      color: Colors.red,
      child: Text('Workflow : $workflow'),
    );
  }
}
