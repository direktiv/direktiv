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
                                  Icon(ExampleConst.breadcrumbsIcon[index]),
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
                        splashColor: ExampleColors.accent,
                        onTap: index == 0
                            ? () {
                                Navigator.popUntil(context,
                                    (Route<dynamic> route) => route.isFirst);
                              }
                            : () {
                                Navigator.pushReplacementNamed(
                                    context, path[key]);
                              },
                        textColor: ExampleColors.primary,
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

class ExampleConst {
  static const List<String> breadcrumbs = [
    'Home',
    'demo-jcmxg',
    'UzUSOV',
  ];
  static const List<IconData> breadcrumbsIcon = [
    Icons.home,
  ];
}

class ExampleColors {
  static const Color primary = Colors.cyan;
  static const Color accent = Colors.indigo;
  static const Color background = Color(0xffEDEDED);

  static const Color primaryTextColor = Colors.white70;
  static const Color secondaryTextColor = Colors.black87;
  static const Color greyTextColor = Colors.black38;
}

class NavBar extends StatelessWidget {
  NavBar({@required this.path});
  final double margin = 10;
  final Map<String, String> path;

  @override
  Widget build(BuildContext context) {
    return Container(
        margin: EdgeInsets.only(
            top: margin, bottom: margin, left: margin, right: margin),
        child: BreadCrumb.builder(
          itemCount: path.length,
          builder: (index) {
            String key = path.keys.elementAt(index);
            String goPath = "/" + path[key];
            return (BreadCrumbItem(
              content: index != 0
                  ? Text(
                      key,
                      style: TextStyle(
                        fontWeight:
                            index < 3 ? FontWeight.bold : FontWeight.normal,
                      ),
                    )
                  : Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(ExampleConst.breadcrumbsIcon[index]),
                        Text(
                          key,
                          style: TextStyle(
                            fontWeight:
                                index < 3 ? FontWeight.bold : FontWeight.normal,
                          ),
                        ),
                      ],
                    ),
              borderRadius: BorderRadius.circular(4),
              padding: EdgeInsets.only(
                  top: margin, bottom: margin, left: margin, right: 4),
              splashColor: ExampleColors.accent,
              onTap: index == 0
                  ? () {
                      Navigator.popUntil(
                          context, (Route<dynamic> route) => route.isFirst);
                    }
                  : Navigator.pushReplacementNamed(context, goPath),
              textColor: ExampleColors.primary,
              disabledTextColor: Colors.grey,
            ));
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
        ));
  }
}
