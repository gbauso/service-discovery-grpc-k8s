from diagrams import Cluster, Diagram, Edge
from diagrams.custom import Custom
from diagrams.programming.language import Go

with Diagram("Service Discovery", show=False):
    client = Custom("Client", "./resources/grpc-icon-black.png")

    with Cluster("K8S Cluster"):
        master = Go("Discovery Master")
        with Cluster("Pod"):
            agent = Go("Discovery Agent") 
            service = Custom("Service", "./resources/grpc-icon-black.png")

    client >> Edge(label="3") >> master << Edge(label="2") << agent >> Edge(label="1") >> service 