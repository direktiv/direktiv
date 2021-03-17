import 'package:flutter/widgets.dart';
import 'package:flutter/material.dart';
import 'package:readonlyui/globals.dart';

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
    workflows = fetchWorkflow(widget.namespace);
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
              child: Text("Workflows"),
            ),
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
                                height: 50,
                                child: OutlinedButton(
                                  style: OutlinedButton.styleFrom(
                                      alignment: Alignment.centerLeft,
                                      padding: EdgeInsets.all(0)),
                                  child: Padding(
                                      padding: EdgeInsets.only(left: 16),
                                      child:
                                          Text('${snapshot.data[index].id}')),
                                  onPressed: () {
                                    print('Pressed ${snapshot.data[index].id}');
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
