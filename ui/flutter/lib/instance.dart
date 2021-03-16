import 'package:flutter/widgets.dart';
import 'package:flutter/material.dart';

class InstanceDetails extends StatelessWidget {
  InstanceDetails({@required this.instance});
  final String instance;

  @override
  Widget build(BuildContext context) {
    return Container(
      color: Colors.red,
      child: Text('Instance Page: $instance'),
    );
  }
}
