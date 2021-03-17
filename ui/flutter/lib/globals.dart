import 'dart:convert';
import 'package:http/http.dart' as http;

const API_SERVER = String.fromEnvironment('API_SERVER', defaultValue: '');

class Workflow {
  final String uid;
  final String id;
  final int revision;
  final bool active;
  final String createdAt;
  final String description;
  final String data;

  Workflow(this.uid, this.id, this.revision, this.active, this.createdAt,
      this.description, this.data);

  factory Workflow.fromJson(Map<String, dynamic> json) {
    String createdAt = "";
    String data;

    if (json.containsKey("workflow")) {
      data = utf8.decode(base64.decode(json["workflow"]));
    }

    return new Workflow(json["uid"], json["id"], json["revision"],
        json["active"], createdAt, json["description"], data);
  }
}

Future<String> executeWorkflow(String namespace, String workflow) async {
  final response = await http.post(
      '$API_SERVER/api/namespaces/$namespace/workflows/$workflow/execute');
  if (response.statusCode == 200) {
    final Map<String, dynamic> resp = jsonDecode(response.body);
    if (resp.containsKey("instanceId")) {
      return resp["instanceId"].toString();
    }
    return "No ID";
  } else {
    throw Exception('Failed execute workflow');
  }
}

Future<Workflow> fetchWorkflow(String namespace, String workflow) async {
  final response = await http
      .get('$API_SERVER/api/namespaces/$namespace/workflows/$workflow');
  if (response.statusCode == 200) {
    return Workflow.fromJson(jsonDecode(response.body));
  } else {
    throw Exception('Failed to load workflow');
  }
}

Future<List<Workflow>> fetchWorkflows(String namespace) async {
  final response =
      await http.get('$API_SERVER/api/namespaces/$namespace/workflows');
  if (response.statusCode == 200) {
    return jsonStrToWorkflowList(response.body);
  } else {
    throw Exception('Failed to load Workflow List');
  }
}

List<Workflow> jsonStrToWorkflowList(String jsonString) {
  Map<String, dynamic> jsonData = jsonDecode(jsonString);
  List<Map<String, dynamic>> workflows = [];

  if (jsonData.containsKey('workflows')) {
    workflows = jsonData['workflows'].cast<Map<String, dynamic>>();
  }

  List<Workflow> workflowList = [];
  workflows.forEach((element) {
    workflowList.add(Workflow.fromJson(element));
  });
  return workflowList;
}

Future<List<Instance>> fetchNamespaceInstances(String namespace) async {
  final response = await http.get('$API_SERVER/api/instances/$namespace');
  if (response.statusCode == 200) {
    return jsonStrToInstanceList(response.body);
  } else {
    throw Exception('Failed to load Instance List');
  }
}

List<Instance> jsonStrToInstanceList(String jsonString) {
  Map<String, dynamic> jsonData = jsonDecode(jsonString);
  List<Map<String, dynamic>> instances = [];

  if (jsonData.containsKey('workflowInstances')) {
    instances = jsonData['workflowInstances'].cast<Map<String, dynamic>>();
  }

  List<Instance> instanceList = [];
  instances.forEach((element) {
    instanceList.add(Instance.fromJson(element));
  });
  return instanceList;
}

class Instance {
  final String workflow;
  final String namespace;
  final String instanceID;
  final String status;

  Instance(this.workflow, this.namespace, this.instanceID, this.status);

  factory Instance.fromJson(Map<String, dynamic> json) {
    String wfInstance = json["id"].toString();
    List<String> wfiSplit = wfInstance.split("/");
    return new Instance(wfiSplit[0], wfiSplit[1], wfiSplit[2], json["status"]);
  }

  Map<String, dynamic> toJson() => {
        'id': '$namespace/$workflow/$instanceID',
        'status': status,
      };
}

class Namespace {
  String name;
  List<Instance> instances;
  Namespace(this.name, this.instances);
}

Future<List<Namespace>> fetchNamespaces() async {
  final response = await http.get('$API_SERVER/api/namespaces');
  List<Namespace> namespaceList = [];
  if (response.statusCode == 200) {
    final nsList = jsonStrToNamespaceList(response.body);
    for (var ns in nsList) {
      final nsInstances = await fetchNamespaceInstances(ns);
      namespaceList.add(Namespace(ns, nsInstances));
    }
    return namespaceList;
  } else {
    throw Exception('Failed to load Namespace List');
  }
}

List<String> jsonStrToNamespaceList(String jsonString) {
  Map<String, dynamic> jsonData = jsonDecode(jsonString);
  final List<Map<String, dynamic>> ns =
      jsonData['namespaces'].cast<Map<String, dynamic>>();
  List<String> namespaceList = [];
  ns.forEach((element) {
    namespaceList.add(element['name']);
  });
  return namespaceList;
}
