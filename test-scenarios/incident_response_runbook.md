# Payment Service Incident Response Runbook

## 🚨 Severity Classification

### **P0 - Critical** (Complete service outage)
- Payment processing completely down
- Database connectivity lost
- Security breach detected
- **Response Time**: 15 minutes
- **Escalation**: Immediate to VP Engineering

### **P1 - High** (Partial service degradation)
- High error rates (>5%)
- Slow response times (>3s)
- Single payment method failing
- **Response Time**: 30 minutes
- **Escalation**: Engineering Manager within 1 hour

### **P2 - Medium** (Minor issues)
- Isolated payment failures
- Non-critical feature issues
- Performance degradation <20%
- **Response Time**: 2 hours
- **Escalation**: Next business day

## 🔍 Initial Assessment Checklist

### 1. **Immediate Health Check**
```bash
# Check pod status
kubectl get pods -n production -l app=payment-service

# Check recent logs
kubectl logs -n production -l app=payment-service --tail=100

# Check service endpoints
kubectl get svc -n production payment-service
```

### 2. **Quick Metrics Review**
- **Error Rate**: Should be <1%
- **Response Time**: Should be <500ms
- **Throughput**: Normal baseline ~1000 req/min
- **Memory Usage**: Should be <80% of limit
- **CPU Usage**: Should be <70% of limit

### 3. **External Dependencies**
- [ ] Database connectivity (PostgreSQL)
- [ ] Redis cache availability
- [ ] External payment gateways (Stripe, PayPal)
- [ ] Authentication service status

## 🛠️ Common Issues & Solutions

### **Pod CrashLoopBackOff**
**Symptoms**: Pod continuously restarting
**Quick Fix**:
1. Check recent deployments: `kubectl rollout history deployment/payment-service -n production`
2. Rollback if needed: `kubectl rollout undo deployment/payment-service -n production`
3. Check resource limits: `kubectl describe pod <pod-name> -n production`

### **High Memory Usage**
**Symptoms**: Memory usage >90%, potential OOMKilled
**Quick Fix**:
1. Restart pods: `kubectl delete pods -n production -l app=payment-service`
2. Check for memory leaks in recent code changes
3. Scale up if needed: `kubectl scale deployment payment-service --replicas=5 -n production`

### **Database Connection Issues**
**Symptoms**: Connection timeout errors in logs
**Quick Fix**:
1. Check database status: `kubectl get pods -n database -l app=postgresql`
2. Verify connection pool settings
3. Check network policies: `kubectl get networkpolicies -n production`

### **Payment Gateway Failures**
**Symptoms**: External API errors, timeout responses
**Quick Fix**:
1. Check gateway status pages (Stripe, PayPal)
2. Verify API credentials and rate limits
3. Enable fallback payment methods
4. Contact vendor support if widespread

## 📞 Communication Protocol

### **Incident Declaration**
1. **Slack**: Post in `#incidents` channel
   ```
   🚨 INCIDENT DECLARED - Payment Service
   Severity: P0/P1/P2
   Impact: [Brief description]
   Incident Commander: @username
   Status Page: Updated/Updating
   ```

2. **Status Page**: Update within 5 minutes
3. **Stakeholder Notification**: 
   - P0: Immediate (CEO, CTO, VP Engineering)
   - P1: Within 30 minutes (Engineering Manager, Product Manager)
   - P2: Next business day (Team Lead)

### **Communication Updates**
- **P0**: Every 15 minutes
- **P1**: Every 30 minutes  
- **P2**: Every 2 hours

## 🔄 Standard Recovery Procedures

### **Rollback Procedure**
1. Identify last known good version:
   ```bash
   kubectl rollout history deployment/payment-service -n production
   ```

2. Execute rollback:
   ```bash
   kubectl rollout undo deployment/payment-service -n production
   ```

3. Verify rollback success:
   ```bash
   kubectl rollout status deployment/payment-service -n production
   ```

4. Monitor for 15 minutes post-rollback

### **Scaling Procedure**
1. Current replica count:
   ```bash
   kubectl get deployment payment-service -n production
   ```

2. Scale up (if needed):
   ```bash
   kubectl scale deployment payment-service --replicas=8 -n production
   ```

3. Monitor resource usage and performance

### **Configuration Reload**
1. Update ConfigMap:
   ```bash
   kubectl edit configmap payment-service-config -n production
   ```

2. Restart deployment:
   ```bash
   kubectl rollout restart deployment/payment-service -n production
   ```

## 📊 Monitoring & Alerting

### **Key Dashboards**
- **Grafana**: Payment Service Overview
- **Datadog**: Application Performance
- **Kubernetes Dashboard**: Cluster Health
- **Sentry**: Error Tracking

### **Critical Alerts**
- Payment success rate <95%
- Response time >2 seconds
- Pod restart count >3 in 10 minutes
- Memory usage >85%
- Database connection failures

## 📝 Post-Incident Actions

### **Immediate (Within 2 hours)**
- [ ] Service fully restored
- [ ] Incident channel archived
- [ ] Status page updated to "resolved"
- [ ] Initial incident summary shared

### **Within 24 Hours**
- [ ] Detailed postmortem scheduled
- [ ] Root cause analysis completed
- [ ] Action items identified and assigned
- [ ] Stakeholder communication sent

### **Within 1 Week**
- [ ] Postmortem document published
- [ ] Preventive measures implemented
- [ ] Monitoring/alerting improvements deployed
- [ ] Runbook updated with lessons learned

## 🎯 Prevention Strategies

### **Code Quality**
- Mandatory code reviews for payment logic
- Load testing before production deployment
- Feature flags for gradual rollout
- Automated rollback triggers

### **Infrastructure**
- Auto-scaling based on traffic patterns
- Circuit breaker patterns for external APIs
- Database connection pooling optimization
- Regular disaster recovery testing

### **Monitoring**
- Synthetic transaction monitoring
- Real user monitoring (RUM)
- Log aggregation and analysis
- Performance regression detection

## 📋 Team Contacts

### **On-Call Rotation**
- **Primary**: Check PagerDuty schedule
- **Secondary**: Check Slack /oncall command
- **Escalation**: VP Engineering (Slack: @vp-eng)

### **Vendor Contacts**
- **Stripe Support**: [Priority support number]
- **PayPal Technical**: [24/7 support line]
- **AWS Support**: [Enterprise support case]
- **Database Vendor**: [Critical issue hotline]

## 🔧 Useful Commands Reference

### **Kubernetes**
```bash
# Get pod status
kubectl get pods -n production -l app=payment-service

# Check logs
kubectl logs -f deployment/payment-service -n production

# Describe deployment
kubectl describe deployment payment-service -n production

# Check events
kubectl get events -n production --sort-by='.lastTimestamp'
```

### **Database**
```bash
# Check connection
kubectl exec -it postgresql-0 -n database -- psql -U postgres -c "SELECT 1"

# Check active connections
kubectl exec -it postgresql-0 -n database -- psql -U postgres -c "SELECT count(*) FROM pg_stat_activity"
```

### **Monitoring**
```bash
# Port forward to access local metrics
kubectl port-forward svc/payment-service-metrics 8080:8080 -n production

# Check service mesh metrics (if using Istio)
kubectl exec -it deployment/payment-service -n production -- curl localhost:15000/stats
```

---

**Last Updated**: [Current Date]  
**Version**: 2.1  
**Next Review**: [30 days from last update]

> 💡 **Remember**: When in doubt, prioritize customer impact mitigation over root cause analysis. Fix first, investigate second.