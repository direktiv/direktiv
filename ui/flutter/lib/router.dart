import 'package:fluro/fluro.dart';
import 'package:flutter/painting.dart';
import 'package:fluro/fluro.dart';
import 'package:flutter/material.dart';

import 'home.dart';
import 'instance.dart';
import 'namespace.dart';
import 'nav.dart';
import 'workflow.dart';

class Application {
  static FluroRouter router;
}

var rootHandler =
    Handler(handlerFunc: (BuildContext context, Map<String, dynamic> params) {
  print(context);
  return NavWrapper(path: {"Home": "/"}, child: Home());
});

var instanceDetailHandler =
    Handler(handlerFunc: (BuildContext context, Map<String, dynamic> params) {
  return NavWrapper(
      path: {"Home": "/"},
      child: InstanceDetails(
        instance: params["id"][0],
      ));
});

var namespaceHandler =
    Handler(handlerFunc: (BuildContext context, Map<String, dynamic> params) {
  final String namespace = params["namespace"][0];

  return NavWrapper(
      path: {"Home": "/", namespace: "/p/$namespace"},
      child: NamespaceWorkflowList(
        namespace: namespace,
      ));
});

var workflowHandler =
    Handler(handlerFunc: (BuildContext context, Map<String, dynamic> params) {
  final String namespace = params["namespace"][0];
  final String workflow = params["workflow"][0];

  return NavWrapper(
      path: {"Home": "/", namespace: "/p/$namespace", workflow: "$workflow"},
      child: Workflow(
        workflow: workflow,
      ));
});

class Routes {
  static String root = "/";
  static String instanceDetails = "/instance/:id";
  static String namespaceWorkflows = "/p/:namespace";
  static String workflowDetails = "/p/:namespace/w/:workflow";

  static void configureRoutes(FluroRouter router) {
    router.notFoundHandler = Handler(
        handlerFunc: (BuildContext context, Map<String, List<String>> params) {
      print("ROUTE WAS NOT FOUND !!!");
      return;
    });
    router.define(root,
        handler: rootHandler, transitionDuration: Duration(milliseconds: 0));
    router.define(instanceDetails,
        handler: instanceDetailHandler,
        transitionDuration: Duration(milliseconds: 0));
    router.define(namespaceWorkflows,
        handler: namespaceHandler,
        transitionDuration: Duration(milliseconds: 0));
    router.define(workflowDetails,
        handler: workflowHandler,
        transitionDuration: Duration(milliseconds: 0));
  }
}
