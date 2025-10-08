# 🔧 Troubleshooting Guide

## Docker Compose Issues

### Service Won't Start
```bash
# Check logs
docker-compose logs dast-api

# Verify all services
docker-compose ps
```

### Database Connection Issues
```bash
# Test database connectivity
docker-compose exec dast-db mysql -u user -ppassword -e "SELECT 1"
```

### ZAP Scanner Not Responding
```bash
# Check ZAP health
curl -H "X-ZAP-API-Key: change-me-9203935709" \
     http://localhost:8090/JSON/core/view/version/
```

## Kubernetes Issues

### Pods Not Starting
```bash
# Check pod status
kubectl get pods -n dast-orchestrator

# Check events
kubectl describe pod <pod-name> -n dast-orchestrator

# Check logs
kubectl logs -f deployment/dast-orchestrator -n dast-orchestrator -c dast-api
```

### Image Pull Issues
```bash
# Verify images exist
docker images | grep dast

# For local clusters, push to local registry
docker push localhost:5000/dast-api:latest
```

### Service Not Accessible
```bash
# Check services
kubectl get svc -n dast-orchestrator

# Port forward for testing
kubectl port-forward svc/dast-api-service 8080:8080 -n dast-orchestrator

# Test health
curl http://localhost:8080/ping
```

## Common Error Messages

### "not connected to zap scanner instance"
- **Cause**: ZAP container not ready or networking issue
- **Solution**: Check ZAP container logs, verify port 8090 is accessible
- **Quick Fix**: Use `/reload/<UUID>` endpoint to reconnect

### "couldn't parse scan information from body"
- **Cause**: Invalid JSON in request body
- **Solution**: Verify request body matches expected format

### "zap client error"
- **Cause**: ZAP scanner internal error
- **Solution**: Check ZAP container logs, restart if needed

## Health Check Responses

### All Services Healthy
```json
{"api":"ok","zap":"ok","dbro":"ok","dbrw":"ok"}
```

### ZAP Connection Failed
```json
{"api":"ok","zap":"failed","dbro":"ok","dbrw":"ok"}
```

### Database Connection Failed
```json
{"api":"ok","zap":"ok","dbro":"failed","dbrw":"failed"}
```

## Performance Issues

### Scans Taking Too Long
- Check `scanMaxDurationMinutes` setting (default: 2 minutes)
- Verify target application is responsive
- Review ZAP scanner resource limits

### High Memory Usage
- Increase container memory limits
- Reduce concurrent scan count
- Clear ZAP sessions after completion
