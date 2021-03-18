import 'package:flutter/widgets.dart';
import 'package:flutter/material.dart';
import 'package:flutter_highlight/flutter_highlight.dart';
import 'package:flutter_highlight/themes/lightfair.dart';
import 'package:readonlyui/router.dart';

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
      child: FutureBuilder<Workflow>(
          future: workflow,
          builder: (BuildContext context, AsyncSnapshot<Workflow> snapshot) {
            if (snapshot.hasData) {
              return (Column(
                children: [
                  OutlinedButton(
                    child: Text('Run'),
                    style: OutlinedButton.styleFrom(
                      primary: Colors.teal,
                    ),
                    onPressed: () {
                      debugPrint('Pressed');
                      executeWorkflow(widget.namespace, widget.workflowID)
                          .then((id) {
                        ScaffoldMessenger.of(context).removeCurrentSnackBar();

                        ScaffoldMessenger.of(context).showSnackBar(SnackBar(
                          action: SnackBarAction(
                            label: 'Open',
                            onPressed: () {
                              Application.router.navigateTo(context, "/i/$id");
                            },
                          ),
                          content: Text("Instance Executed"),
                          duration: const Duration(milliseconds: 3000),
                          width: 220.0, // Width of the SnackBar.
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8.0,
                          ),
                          behavior: SnackBarBehavior.floating,
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(10.0),
                          ),
                        ));
                      });
                    },
                  ),
                  Expanded(
                      child: HighlightView(
                    snapshot.data.data,
                    language: 'yaml',
                    theme: lightfairTheme,
                    padding: EdgeInsets.all(3),
                    textStyle: TextStyle(
                      fontSize: 16,
                    ),
                  ))
                ],
              ));
            } else if (snapshot.hasError) {
              return Text("${snapshot.error}");
            }
            return CircularProgressIndicator();
          }),
    );
  }
}
