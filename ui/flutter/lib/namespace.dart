import 'package:flutter/widgets.dart';
import 'package:flutter/material.dart';

class NamespaceWorkflowList extends StatelessWidget {
  NamespaceWorkflowList({@required this.namespace});
  final String namespace;

  @override
  Widget build(BuildContext context) {
    return Container(
      color: Colors.red,
      child: Text('Namespace : $namespace'),
    );
  }
}
