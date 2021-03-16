import 'package:flutter/rendering.dart';
import 'package:readonlyui/router.dart';
import 'globals.dart' as globals;
import 'package:flutter/material.dart';
import 'instance.dart';
import 'nav.dart';

class Home extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        ConstrainedBox(
          child: NamespaceList(
            title: "Namespaces",
            namespaces: globals.jsonStrToNamespaceList(globals.namespaceList),
          ),
          constraints: BoxConstraints(
            minWidth: 150,
            maxWidth: 220,
          ),
        ),
        Expanded(
            child: InstanceList(
          title: "Instances",
          instances: globals.jsonStrToInstanceList(globals.instanceList),
        ))
      ],
    );
  }
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
            Container(
              padding: const EdgeInsets.only(left: 16),
              height: 50,
              alignment: Alignment.centerLeft,
              child: Text(title),
            ),
            Expanded(
                child: ListView.builder(
                    padding: const EdgeInsets.all(0),
                    itemCount: instances.length,
                    itemBuilder: (BuildContext context, int index) {
                      return Container(
                        height: 50,
                        child: OutlinedButton(
                          style: OutlinedButton.styleFrom(
                              alignment: Alignment.centerLeft,
                              padding: EdgeInsets.all(0)),
                          child: Padding(
                              padding: EdgeInsets.only(left: 16),
                              child: Text('${instances[index].instanceID}')),
                          onPressed: () {
                            print('Pressed ${instances[index].instanceID}');
                            Application.router.navigateTo(context,
                                "/instance/${instances[index].instanceID}");
                          },
                        ),
                      );
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

class NamespaceList extends StatelessWidget {
  NamespaceList({@required this.namespaces, @required this.title});
  final List<String> namespaces;
  final String title;
  final double margin = 10;

  @override
  Widget build(BuildContext context) {
    return Container(
        margin: EdgeInsets.only(
            top: margin, bottom: margin, left: margin, right: 0),
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.only(left: 16),
              height: 50,
              alignment: Alignment.centerLeft,
              child: Text(title),
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
                              padding: EdgeInsets.all(0)),
                          child: Padding(
                              padding: EdgeInsets.only(left: 16),
                              child: Text('${namespaces[index]}')),
                          onPressed: () {
                            print('Pressed ${namespaces[index]}');
                            Application.router
                                .navigateTo(context, "/p/${namespaces[index]}");
                          },
                        ),
                      );
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
