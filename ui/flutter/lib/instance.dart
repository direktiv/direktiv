import 'package:flutter/widgets.dart';
import 'package:flutter/material.dart';
import 'package:readonlyui/globals.dart';

import 'logs.dart';

class InstanceDetails extends StatefulWidget {
  InstanceDetails({@required this.instanceID});
  final String instanceID;

  @override
  _InstanceDetailsState createState() => _InstanceDetailsState();
}

class _InstanceDetailsState extends State<InstanceDetails> {
  Future<InstanceDetail> instance;

  @override
  void initState() {
    super.initState();
    instance = fetchInstanceDetail(widget.instanceID);
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      child: FutureBuilder<InstanceDetail>(
          future: instance,
          builder:
              (BuildContext context, AsyncSnapshot<InstanceDetail> snapshot) {
            if (snapshot.hasData) {
              return (Column(
                children: [
                  Text("${snapshot.data.input}"),
                  Expanded(
                      child: InstanceLogPage(instanceID: widget.instanceID))
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
