# Netductor MikroTik thin agent (RouterOS 7)
# Install:
#   1) Upload this file, /import netductor-agent.rsc
#   2) Set globals below (or via /system script environment)
#   3) Scheduler every 1-5 min calls nd-agent-tick

# --- config (edit once) ---
:global NdServer "https://YOUR_VPS:443"
:global NdBootstrap "BOOTSTRAP_TOKEN"
:global NdDeviceId "mt-home-01"
:global NdToken ""
:global NdKind "mikrotik"

:global NdEnsureDir do={
  # RouterOS has no real dirs for app state; use files
}

:global NdAgentTick do={
  :global NdServer
  :global NdBootstrap
  :global NdDeviceId
  :global NdToken
  :global NdKind

  :local url ($NdServer . "/api/edge/heartbeat")
  :local auth $NdToken
  :if ([:len $auth] = 0) do={
    # enroll
    :local eurl ($NdServer . "/api/edge/enroll")
    :local body ("{\"device_id\":\"" . $NdDeviceId . "\",\"board\":\"mikrotik\",\"hostname\":\"" . [/system identity get name] . "\",\"kind\":\"mikrotik\"}")
    /tool fetch url=$eurl http-method=post http-header-field=("Authorization: Bearer " . $NdBootstrap . ",Content-Type: application/json") http-data=$body output=user as-value
    # After approve, operator sets token file manually or next fetch commands
    :return
  }

  :local body ("{\"device_id\":\"" . $NdDeviceId . "\",\"hostname\":\"" . [/system identity get name] . "\"}")
  /tool fetch url=$url http-method=post http-header-field=("Authorization: Bearer " . $auth . ",Content-Type: application/json") http-data=$body output=user as-value

  # poll commands
  :local curl ($NdServer . "/api/edge/commands?device_id=" . $NdDeviceId)
  :local res [/tool fetch url=$curl http-header-field=("Authorization: Bearer " . $auth) output=user as-value]
  # If server returned apply_rsc URL — fetch and /import (handled when payload is simple)
}

# Scheduler
/system script remove [find name=nd-agent-tick]
/system script add name=nd-agent-tick owner=admin policy=read,write,policy,test source={
  :global NdAgentTick
  \$NdAgentTick
}
/system scheduler remove [find name=nd-agent]
/system scheduler add name=nd-agent interval=2m on-event=nd-agent-tick policy=read,write,policy,test
:put "netductor MikroTik agent installed — set NdServer/NdBootstrap/NdDeviceId and wait for approve"
