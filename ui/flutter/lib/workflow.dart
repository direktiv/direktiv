import 'package:flutter/gestures.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter/material.dart';
import 'package:flutter_highlight/flutter_highlight.dart';
import 'package:flutter_highlight/themes/lightfair.dart';
import 'package:readonlyui/router.dart';
import 'package:readonlyui/themes.dart';

import 'globals.dart';
import 'home.dart';

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
              return (ConstrainedBox(
                  constraints: BoxConstraints(
                      maxHeight: 800, minHeight: 300, minWidth: 300),
                  child: Column(
                    children: [
                      FractionallySizedBox(
                          alignment: Alignment.topCenter,
                          widthFactor: 0.8,

                          child: Container(child: Row(
                            children: [
                              Text(widget.workflowID, style: new TextStyle(
                              color: Colors.black,
                              fontSize: 20,
                              fontWeight: FontWeight.bold)),
                              Expanded(child: Container(
                                  alignment: Alignment.centerRight,
                                  child: OutlinedButton(
                                    child: Text('Run'),
                                    style: OutlinedButton.styleFrom(
                                      primary: Colors.white,
                                      backgroundColor: colors.status("complete"),
                                    ),
                                    onPressed: () {
                                      _newTaskModalBottomSheet(context,
                                          widget.namespace, widget.workflowID);
                                    },
                                  )))
                            ],
                          ),
                              padding: EdgeInsets.only(bottom: 10))),
                      Expanded(
                          child: FractionallySizedBox(
                              alignment: Alignment.topCenter,
                              widthFactor: 0.8,
                              heightFactor: 0.7,
                              child: Container(
                                  padding: EdgeInsets.all(5),
                                  child: HighlightView(
                                    snapshot.data.data,
                                    language: 'yaml',
                                    theme: lightfairTheme,
                                    padding: EdgeInsets.all(3),
                                    textStyle: TextStyle(
                                      fontSize: 16,
                                    ),
                                  ),
                                  decoration: BoxDecoration(
                                      color: Colors.white,
                                      border: Border.all(
                                        color: colors.background(2),
                                        width: 1,
                                      ),
                                      borderRadius: BorderRadius.all(
                                          Radius.circular(5))))))
                    ],
                  )));
            } else if (snapshot.hasError) {
              return Text("${snapshot.error}");
            }
            return CircularProgressIndicator();
          }),
    );
  }
}

void _newTaskModalBottomSheet(context, String namespace, String workflowID) {
  showModalBottomSheet(
      context: context,
      builder: (BuildContext bc) {
        return Container(
          child: new Wrap(
            children: <Widget>[
              new MyCustomForm(namespace: namespace, workflowID: workflowID),
            ],
          ),
        );
      });
}

// Define a custom Form widget.
class MyCustomForm extends StatefulWidget {
  MyCustomForm({@required this.namespace, @required this.workflowID});

  final String namespace;
  final String workflowID;

  @override
  _MyCustomFormState createState() => _MyCustomFormState();
}

// Define a corresponding State class.
// This class holds the data related to the Form.
class _MyCustomFormState extends State<MyCustomForm> {
  // Create a text controller and use it to retrieve the current value
  // of the TextField.
  final myController = TextEditingController();
  var error;
  var instID;

  @override
  void dispose() {
    myController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Container(
        padding: EdgeInsets.all(10),
        child: ConstrainedBox(
            constraints: BoxConstraints(minHeight: 300, minWidth: 300),
            child: Column(children: [
              Container(
                  alignment: Alignment.centerLeft,
                  margin: EdgeInsets.only(bottom: 10),
                  child: Text(
                    "Input",
                    style: new TextStyle(
                        color: Colors.black,
                        fontSize: 20,
                        fontWeight: FontWeight.bold),
                  ),
                  decoration: BoxDecoration(
                    border: Border(
                        bottom:
                            BorderSide(width: 1, color: colors.background(2))),
                  )),
              Row(
                children: [
                  Container(
                      alignment: Alignment.centerLeft,
                      child: (() {
                    if (error != null) {
                      return TextBoldPrefix("Error: ", error, color: colors.status("failed"));
                    } else if (instID != null) {
                      return new RichText(
                        text: new TextSpan(
                          children: [
                            new TextSpan(
                              text: 'Instance Created: ',
                              style: new TextStyle(
                                  color: Colors.black,
                                  fontWeight: FontWeight.bold),
                            ),
                            new TextSpan(
                              text: instID,
                              style: new TextStyle(color: Colors.blue),
                              recognizer: new TapGestureRecognizer()
                                ..onTap = () {
                                  Application.router
                                      .navigateTo(context, "/i/$instID");
                                },
                            ),
                          ],
                        ),
                      );
                      return TextBoldPrefix("Instance Created: ", instID);
                    } else {
                      return SizedBox.shrink();
                    }
                  }())),
                  Expanded(
                      child: Container(
                          padding: EdgeInsets.only(bottom: 10),
                          alignment: Alignment.centerRight,
                          child: OutlinedButton(
                            child: Text('Run'),
                            style: OutlinedButton.styleFrom(
                              primary: Colors.white,
                              backgroundColor: colors.status("complete"),
                            ),
                            onPressed: () {
                              executeWorkflow(
                                      widget.namespace, widget.workflowID)
                                  .then((id) {
                                setState(() {
                                  error = null;
                                  instID = id;
                                });
                              }).catchError((e) {
                                setState(() {
                                  error = e.toString();
                                  instID = null;
                                });
                              });
                            },
                          )))
                ],
              ),
              Container(
                  height: 300,
                  padding: EdgeInsets.all(10),
                  child: TextField(
                    maxLines: null,
                    decoration: InputDecoration.collapsed(
                      border: InputBorder.none,
                    ),
                    controller: myController,
                  ),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    border: Border.all(
                      color: colors.background(2),
                      width: 1,
                    ),
                    borderRadius: BorderRadius.circular(3),
                  )),
            ])));
  }
}

// executeWorkflow(widget.namespace, widget.workflowID).then((id) {
//   ScaffoldMessenger.of(context).removeCurrentSnackBar();
//
//   ScaffoldMessenger.of(context).showSnackBar(SnackBar(
//     action: SnackBarAction(
//       label: 'Open',
//       onPressed: () {
//         Application.router.navigateTo(context, "/i/$id");
//       },
//     ),
//     content: Text("Instance Executed"),
//     duration: const Duration(milliseconds: 3000),
//     width: 220.0,
//     // Width of the SnackBar.
//     padding: const EdgeInsets.symmetric(
//       horizontal: 8.0,
//     ),
//     behavior: SnackBarBehavior.floating,
//     shape: RoundedRectangleBorder(
//       borderRadius: BorderRadius.circular(10.0),
//     ),
//   ));
// });
// Navigator.pop(context);
