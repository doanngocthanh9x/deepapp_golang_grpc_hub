#!/bin/bash
# View logs from all services in all-in-one container

case "$1" in
  hub)
    echo "📋 Hub logs:"
    docker exec deepapp-hub-all-in-one tail -f /var/log/supervisor/hub.out.log
    ;;
  webapi)
    echo "📋 Web API logs:"
    docker exec deepapp-hub-all-in-one tail -f /var/log/supervisor/webapi.out.log
    ;;
  java)
    echo "📋 Java Worker logs:"
    docker exec deepapp-hub-all-in-one tail -f /var/log/supervisor/java-worker.out.log
    ;;
  python)
    echo "📋 Python Worker logs:"
    docker exec deepapp-hub-all-in-one tail -f /var/log/supervisor/python-worker.out.log
    ;;
  status)
    echo "📊 Service status:"
    docker exec deepapp-hub-all-in-one supervisorctl status
    ;;
  all)
    echo "📋 All services logs (last 50 lines each):"
    echo ""
    echo "=== HUB ==="
    docker exec deepapp-hub-all-in-one tail -20 /var/log/supervisor/hub.out.log
    echo ""
    echo "=== WEB API ==="
    docker exec deepapp-hub-all-in-one tail -20 /var/log/supervisor/webapi.out.log
    echo ""
    echo "=== JAVA WORKER ==="
    docker exec deepapp-hub-all-in-one tail -20 /var/log/supervisor/java-worker.out.log
    echo ""
    echo "=== PYTHON WORKER ==="
    docker exec deepapp-hub-all-in-one tail -20 /var/log/supervisor/python-worker.out.log
    ;;
  *)
    echo "Usage: $0 {hub|webapi|java|python|status|all}"
    echo ""
    echo "Examples:"
    echo "  $0 hub       - View Hub logs (live)"
    echo "  $0 webapi    - View Web API logs (live)"
    echo "  $0 java      - View Java Worker logs (live)"
    echo "  $0 python    - View Python Worker logs (live)"
    echo "  $0 status    - View all service status"
    echo "  $0 all       - View last 20 lines from all services"
    exit 1
    ;;
esac
