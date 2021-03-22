import 'package:flutter/rendering.dart';
import 'package:flutter/material.dart';
import 'package:flutter_breadcrumb/flutter_breadcrumb.dart';
import "router.dart";

class NavWrapper extends StatelessWidget {
  NavWrapper({@required this.child, @required this.path});
  final Widget child;
  final Map<String, String> path;
  final double margin = 10;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
        appBar: AppBar(
          title: Center(child: Text('Direktiv')),
          flexibleSpace: Container(
            decoration: new BoxDecoration(
              gradient: new LinearGradient(
                  colors: [
                    const Color(0xFF0061EB),
                    const Color(0xFF009FFB),
                  ],
                  begin: const FractionalOffset(0.0, 0.0),
                  end: const FractionalOffset(1.0, 0.0),
                  stops: [0.0, 1.0],
                  tileMode: TileMode.clamp),
            ),
          ),
        ),
        body: Container(
            child: Column(
          children: [
            FractionallySizedBox(
              widthFactor: 1,
              child: Container(
                  margin: EdgeInsets.only(
                      top: margin, bottom: margin, left: margin, right: margin),
                  child: BreadCrumb.builder(
                    itemCount: path.length,
                    builder: (index) {
                      String key = path.keys.elementAt(index);
                      return BreadCrumbItem(
                        content: index != 0
                            ? Text(
                                key,
                                style: TextStyle(
                                  fontWeight: index < 3
                                      ? FontWeight.bold
                                      : FontWeight.normal,
                                ),
                              )
                            : Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Icon(Icons.home),
                                  Text(
                                    key,
                                    style: TextStyle(
                                      fontWeight: index < 3
                                          ? FontWeight.bold
                                          : FontWeight.normal,
                                    ),
                                  ),
                                ],
                              ),
                        borderRadius: BorderRadius.circular(4),
                        padding: EdgeInsets.only(
                            top: margin,
                            bottom: margin,
                            left: margin,
                            right: 4),
                        splashColor: Colors.indigo,
                        onTap: index == 0
                            ? () {
                                Application.router.navigateTo(context, "/");
                              }
                            : () {
                                Application.router
                                    .navigateTo(context, path[key]);
                              },
                        textColor: Colors.cyan,
                        disabledTextColor: Colors.grey,
                      );
                    },
                    divider: Icon(
                      Icons.chevron_right,
                      color: Colors.grey,
                    ),
                  ),
                  decoration: BoxDecoration(
                    border: Border.all(
                      color: Colors.black,
                      width: 3,
                    ),
                    borderRadius: BorderRadius.circular(3),
                  )),
            ),
            Expanded(
              child: child,
            )
          ],
        )));
  }
}
