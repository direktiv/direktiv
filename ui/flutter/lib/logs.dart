import 'package:flutter/widgets.dart';
import 'package:flutter/material.dart';
import 'package:readonlyui/globals.dart';

class InstanceLogPage extends StatefulWidget {
  InstanceLogPage({@required this.instanceID});
  final String instanceID;

  @override
  _InstanceLogPageState createState() => _InstanceLogPageState();
}

class _InstanceLogPageState extends State<InstanceLogPage> {
  Future<InstanceLogs> logs;

  @override
  void initState() {
    super.initState();
    logs = fetchInstanceLogs(widget.instanceID);
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      child: FutureBuilder<InstanceLogs>(
          future: logs,
          builder:
              (BuildContext context, AsyncSnapshot<InstanceLogs> snapshot) {
            if (snapshot.hasData) {
              return (ListView.builder(
                  padding: const EdgeInsets.all(0),
                  itemCount: snapshot.data.workflowInstanceLogs.length,
                  itemBuilder: (BuildContext context, int index) {
                    return Container(
                      height: 50,
                      child: Text(
                          "${snapshot.data.workflowInstanceLogs[index].message}"),
                    );
                  }));
            } else if (snapshot.hasError) {
              return Text("${snapshot.error}");
            }
            return CircularProgressIndicator();
          }),
    );
  }
}
