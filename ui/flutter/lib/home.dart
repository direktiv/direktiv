import 'package:flutter/rendering.dart';
import 'package:readonlyui/router.dart';
import 'package:readonlyui/themes.dart';
import 'globals.dart' as globals;
import 'package:flutter/material.dart';
import 'globals.dart';
import 'instance.dart';
import 'nav.dart';

class Home extends StatefulWidget {
  @override
  _HomeState createState() => _HomeState();
}

class _HomeState extends State<Home> {
  Future<List<globals.Namespace>> namespaces;

  @override
  void initState() {
    super.initState();
    namespaces = globals.fetchNamespaces();
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<List<globals.Namespace>>(
        future: namespaces,
        builder: (BuildContext context,
            AsyncSnapshot<List<globals.Namespace>> snapshot) {
          if (snapshot.hasData) {
            return (Row(
              children: [
                ConstrainedBox(
                  child: NamespaceList(
                    title: "Namespaces",
                    namespaces: snapshot.data,
                  ),
                  constraints: BoxConstraints(
                    minWidth: 150,
                    maxWidth: 220,
                  ),
                ),
                Expanded(
                    child: InstanceList(
                  title: "Instances",
                  instances: getAllInstances(snapshot.data),
                ))
              ],
            ));
          } else if (snapshot.hasError) {
            return Text("${snapshot.error}");
          }
          return CircularProgressIndicator();
        });
  }
}

List<Instance> getAllInstances(List<Namespace> namespaces) {
  List<Instance> instances = [];
  for (var ns in namespaces) {
    instances.addAll(ns.instances);
  }
  return instances;
}

class InstanceList extends StatelessWidget {
  InstanceList({@required this.instances, @required this.title});

  final List<globals.Instance> instances;
  final String title;
  final double margin = 10;

  @override
  Widget build(BuildContext context) {
    return Container(
        margin: EdgeInsets.only(
            top: margin, bottom: margin, left: margin, right: margin),
        child: Column(
          children: [
            Padding(
              padding: EdgeInsets.symmetric(horizontal: 6),
              child: Container(
                  padding: const EdgeInsets.only(left: 16),
                  height: 50,
                  alignment: Alignment.centerLeft,
                  child: TextBoldPrefix(title, "", fontSize: 20),
                  decoration: BoxDecoration(
                    border: Border(
                      bottom: BorderSide(width: 1, color: colors.background(2))
                    ),
                  )),
            ),
            Expanded(
                child: ListView.builder(
                    padding: const EdgeInsets.all(0),
                    itemCount: instances.length,
                    itemBuilder: (BuildContext context, int index) {
                      return Container(
                        height: 80,
                        padding:
                            EdgeInsets.symmetric(vertical: 6, horizontal: 6),
                        child: OutlinedButton(
                          style: OutlinedButton.styleFrom(
                            alignment: Alignment.centerLeft,
                            padding: EdgeInsets.all(0),
                            side: BorderSide(
                                width: 1, color: colors.status(instances[index].status)),
                            backgroundColor: colors.background(0),
                            shape: const RoundedRectangleBorder(
                                borderRadius:
                                    BorderRadius.all(Radius.circular(5))),
                          ),
                          child: Container(
                              padding: EdgeInsets.all(10),
                              child: Column(
                                children: [
                                  Flexible(
                                      flex: 1,
                                      fit: FlexFit.tight,
                                      child: Row(
                                        children: [
                                          Flexible(
                                              flex: 1,
                                              fit: FlexFit.tight,
                                              child: Container(
                                                  child: TextBoldPrefix("Workflow: ", instances[index].workflow))),
                                          Flexible(
                                              flex: 1,
                                              fit: FlexFit.tight,
                                              child: Container(
                                                  child: TextBoldPrefix("Namespace: ",instances[index].namespace))),
                                        ],
                                      )),
                                  Flexible(
                                      flex: 1,
                                      fit: FlexFit.tight,
                                      child: Row(
                                        children: [
                                          Flexible(
                                              flex: 1,
                                              fit: FlexFit.tight,
                                              child: Container(
                                                  child: TextBoldPrefix("Instance: ", instances[index].instanceID))),
                                          Flexible(
                                              flex: 1,
                                              fit: FlexFit.tight,
                                              child: Container(
                                                  child: Container(
                                                    child: TextBoldPrefix("Status: ", instances[index].status, color: Colors.black),
                                                      ),
                                                      )),
                                        ],
                                      ))
                                ],
                              )),
                          onPressed: () {
                            print('Pressed ${instances[index].instanceID}');
                            Application.router.navigateTo(context,
                                "/i/${instances[index].namespace}/${instances[index].workflow}/${instances[index].instanceID}");
                          },
                        ),
                      );
                    }))
          ],
        ));
  }
}
// Text('Workflow: ${instances[index].workflow}')))
RichText TextBoldPrefix(String prefix, String text, {fontSize = 14.0, color: Colors.black}) {
  return (
      new RichText(
          text: new TextSpan(
            style: new TextStyle(
              fontSize: fontSize,
              color: color,
            ),
            children: <TextSpan>[
              new TextSpan(text: prefix, style: new TextStyle(fontWeight: FontWeight.bold)),
              new TextSpan(text: text),
            ],
          )
  ));
}

class NamespaceList extends StatelessWidget {
  NamespaceList({@required this.namespaces, @required this.title});

  final List<globals.Namespace> namespaces;
  final String title;
  final double margin = 10;

  @override
  Widget build(BuildContext context) {
    print(namespaces);
    return Container(
        margin: EdgeInsets.only(
            top: margin, bottom: margin, left: margin, right: 0),
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.only(left: 16),
              height: 50,
              alignment: Alignment.centerLeft,
              child: TextBoldPrefix(title, ""),
                decoration: BoxDecoration(
                  color: colors.background(0),
                  border: Border.all(
                    color: colors.background(2),
                    width: 1,
                  ),
                    borderRadius: BorderRadius.only(
                      topRight: Radius.circular(5),
                      topLeft: Radius.circular(5),
                    )
                )
            ),
            Expanded(
                child: ListView.builder(
                    padding: const EdgeInsets.all(0),
                    itemCount: namespaces.length,
                    itemBuilder: (BuildContext context, int index) {
                      return Container(
                        height: 50,
                        child: OutlinedButton(
                          style: OutlinedButton.styleFrom(
                              alignment: Alignment.centerLeft,
                              padding: EdgeInsets.all(0),
                            shape: const ContinuousRectangleBorder(

                            ),
                          ),
                          child: Padding(
                              padding: EdgeInsets.only(left: 16),
                              child: Text('${namespaces[index].name}')),
                          onPressed: () {
                            print('Pressed ${namespaces[index].name}');
                            Application.router.navigateTo(
                                context, "/p/${namespaces[index].name}");
                          },
                        ),
                      );
                    }))
          ],
        ));
  }
}
