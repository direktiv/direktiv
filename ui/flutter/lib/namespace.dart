import 'package:flutter/widgets.dart';
import 'package:flutter/material.dart';
import 'package:readonlyui/globals.dart';
import 'package:readonlyui/router.dart';
import 'package:readonlyui/themes.dart';

class NamespaceWorkflowList extends StatefulWidget {
  NamespaceWorkflowList({@required this.namespace});

  final String namespace;

  @override
  _NamespaceWorkflowListState createState() => _NamespaceWorkflowListState();
}

class _NamespaceWorkflowListState extends State<NamespaceWorkflowList> {
  Future<List<Workflow>> workflows;
  final double margin = 10;

  @override
  void initState() {
    super.initState();
    workflows = fetchWorkflows(widget.namespace);
  }

  @override
  Widget build(BuildContext context) {
    return Container(
        margin: EdgeInsets.only(
            top: margin, bottom: margin, left: margin, right: margin),
        child: Column(
          children: [
            Container(
                padding: const EdgeInsets.only(left: 16),
                height: 50,
                alignment: Alignment.centerLeft,
                child: TextBoldPrefix("Workflows", "", fontSize: 20),
                decoration: BoxDecoration(
                  border: Border(
                      bottom: BorderSide(width: 1, color: colors.background(2))
                  ),
                )),
            Expanded(
                child: FutureBuilder<List<Workflow>>(
                    future: workflows,
                    builder: (BuildContext context,
                        AsyncSnapshot<List<Workflow>> snapshot) {
                      if (snapshot.hasData) {
                        return (ListView.builder(
                            padding: const EdgeInsets.all(0),
                            itemCount: snapshot.data.length,
                            itemBuilder: (BuildContext context, int index) {
                              return Container(
                                height: 80,
                                padding: EdgeInsets.symmetric(
                                    vertical: 6, horizontal: 6),
                                child: OutlinedButton(
                                  style: OutlinedButton.styleFrom(
                                    alignment: Alignment.topLeft,
                                    padding: EdgeInsets.all(0),
                                    side: BorderSide(
                                        width: 1, color: colors.background(2)),
                                    backgroundColor: colors.background(0),
                                    shape: const RoundedRectangleBorder(
                                        borderRadius: BorderRadius.all(
                                            Radius.circular(5))),
                                  ),
                                  child: Container(
                                      padding: EdgeInsets.all(10),
                                      child: Column(
                                        children: [
                                          Flexible(
                                              flex: 1,
                                              fit: FlexFit.tight,
                                              child: Container(
                                                alignment: Alignment.centerLeft,
                                                child: Text(
                                                    snapshot.data[index].id,
                                                    style: TextStyle(
                                                        fontWeight:
                                                            FontWeight.bold,
                                                        color: Colors.black)),
                                              )),
                                          Flexible(
                                              flex: 1,
                                              fit: FlexFit.tight,
                                              child: Container(
                                                alignment: Alignment.centerLeft,
                                                child: snapshot.data[index]
                                                            .description ==
                                                        ""
                                                    ? Text("No Description",
                                                        style: TextStyle(
                                                            color:
                                                                Colors.black))
                                                    : Text(
                                                        snapshot.data[index]
                                                            .description,
                                                        style: TextStyle(
                                                            color:
                                                                Colors.black)),
                                              ))
                                        ],
                                      )),
                                  onPressed: () {
                                    print('Pressed ${snapshot.data[index].id}');
                                    Application.router.navigateTo(context,
                                        "/p/${widget.namespace}/w/${snapshot.data[index].id}");
                                  },
                                ),
                              );
                            }));
                      } else if (snapshot.hasError) {
                        return Text("${snapshot.error}");
                      }
                      return CircularProgressIndicator();
                    }))
          ],
        ),
        decoration: BoxDecoration(
          border: Border.all(
            color: Colors.black,
            width: 3,
          ),
          borderRadius: BorderRadius.circular(3),
        ));
  }
}

RichText TextBoldPrefix(String prefix, String text,
    {fontSize = 14.0, color: Colors.black}) {
  return (new RichText(
      text: new TextSpan(
    style: new TextStyle(
      fontSize: fontSize,
      color: color,
    ),
    children: <TextSpan>[
      new TextSpan(
          text: prefix, style: new TextStyle(fontWeight: FontWeight.bold)),
      new TextSpan(text: text),
    ],
  )));
}
