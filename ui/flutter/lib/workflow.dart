import 'package:flutter/widgets.dart';
import 'package:flutter/material.dart';
import 'package:flutter_highlight/flutter_highlight.dart';
import 'package:flutter_highlight/themes/lightfair.dart';

import 'globals.dart';

class WorkflowPage extends StatefulWidget {
  WorkflowPage({@required this.workflowID, @required this.namespace});
  final String workflowID;
  final String namespace;

  @override
  _WorkflowPageState createState() => _WorkflowPageState();
}

class _WorkflowPageState extends State<WorkflowPage> {
  Future<Workflow> workflow;

  @override
  void initState() {
    super.initState();
    workflow = fetchWorkflow(widget.namespace, widget.workflowID);
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      color: Colors.red,
      child: FutureBuilder<Workflow>(
          future: workflow,
          builder: (BuildContext context, AsyncSnapshot<Workflow> snapshot) {
            if (snapshot.hasData) {
              return (HighlightView(
                snapshot.data.data,
                language: 'yaml',
                theme: lightfairTheme,
                padding: EdgeInsets.all(3),
                textStyle: TextStyle(
                  fontSize: 16,
                ),
              ));
            } else if (snapshot.hasError) {
              return Text("${snapshot.error}");
            }
            return CircularProgressIndicator();
          }),
    );
  }
}
