import 'dart:convert';
import 'package:http/http.dart' as http;

const API_SERVER =
    String.fromEnvironment('API_SERVER', defaultValue: 'http://localhost:8080');

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

// Instance Detail
class InstanceDetail {
  String id;
  String status;
  String invokedBy;
  int revision;
  String input;

  InstanceDetail(
      {this.id, this.status, this.invokedBy, this.revision, this.input});

  InstanceDetail.fromJson(Map<String, dynamic> json) {
    id = json['id'];
    status = json['status'];
    invokedBy = json['invokedBy'];
    revision = json['revision'];
    input = utf8.decode(base64.decode(json["input"]));
  }

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = new Map<String, dynamic>();
    data['id'] = this.id;
    data['status'] = this.status;
    data['invokedBy'] = this.invokedBy;
    data['revision'] = this.revision;
    data['input'] = this.input;
    return data;
  }
}

Future<InstanceDetail> fetchInstanceDetail(String instanceID) async {
  final response = await http.get('$API_SERVER/api/instances/$instanceID');
  if (response.statusCode == 200) {
    return InstanceDetail.fromJson(jsonDecode(response.body));
  } else {
    throw Exception('Failed to load Instance List');
  }
}

// Instance Logs

class InstanceLogs {
  List<WorkflowInstanceLogs> workflowInstanceLogs;
  int offset;
  int limit;

  InstanceLogs({this.workflowInstanceLogs, this.offset, this.limit});

  InstanceLogs.fromJson(Map<String, dynamic> json) {
    if (json['workflowInstanceLogs'] != null) {
      workflowInstanceLogs = [];
      json['workflowInstanceLogs'].forEach((v) {
        workflowInstanceLogs.add(new WorkflowInstanceLogs.fromJson(v));
      });
    }
    offset = json['offset'];
    limit = json['limit'];
  }

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = new Map<String, dynamic>();
    if (this.workflowInstanceLogs != null) {
      data['workflowInstanceLogs'] =
          this.workflowInstanceLogs.map((v) => v.toJson()).toList();
    }
    data['offset'] = this.offset;
    data['limit'] = this.limit;
    return data;
  }
}

class WorkflowInstanceLogs {
  Timestamp timestamp;
  String message;

  WorkflowInstanceLogs({this.timestamp, this.message});

  WorkflowInstanceLogs.fromJson(Map<String, dynamic> json) {
    timestamp = json['timestamp'] != null
        ? new Timestamp.fromJson(json['timestamp'])
        : null;
    message = json['message'];
  }

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = new Map<String, dynamic>();
    if (this.timestamp != null) {
      data['timestamp'] = this.timestamp.toJson();
    }
    data['message'] = this.message;
    return data;
  }
}

class Timestamp {
  int seconds;

  Timestamp({this.seconds});

  Timestamp.fromJson(Map<String, dynamic> json) {
    seconds = json['seconds'];
  }

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> data = new Map<String, dynamic>();
    data['seconds'] = this.seconds;
    return data;
  }
}

Future<InstanceLogs> fetchInstanceLogs(String instanceID) async {
  final response =
      await http.get('$API_SERVER/api/instances/$instanceID/logs?limit=1000');
  if (response.statusCode == 200) {
    return InstanceLogs.fromJson(jsonDecode(response.body));
  } else {
    throw Exception('Failed to load Instance List');
  }
}

// Instance Simple

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

  Instance(this.namespace, this.workflow, this.instanceID, this.status);

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
