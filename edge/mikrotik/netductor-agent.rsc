# Netductor MikroTik thin agent (RouterOS 7+) — scheduler + fetch
# Same control plane as OpenWrt agent: enroll → approve → heartbeat → commands
#
# Setup:
#   :global NdServer "https://IP-or-host"
#   :global NdBootstrap "BOOTSTRAP_TOKEN"
#   :global NdDeviceId "mt-home-01"
#   :global NdToken ""
#   /import netductor-agent.rsc

:global NdServer
:global NdBootstrap
:global NdDeviceId
:global NdToken

:if ([:typeof $NdServer] = "nothing") do={ :global NdServer "https://CHANGE_ME" }
:if ([:typeof $NdBootstrap] = "nothing") do={ :global NdBootstrap "" }
:if ([:typeof $NdDeviceId] = "nothing") do={ :global NdDeviceId [/system identity get name] }
:if ([:typeof $NdToken] = "nothing") do={ :global NdToken "" }

:global NdAgentTick do={
  :global NdServer
  :global NdBootstrap
  :global NdDeviceId
  :global NdToken

  :local hn [/system identity get name]
  :local body ("{\"device_id\":\"" . $NdDeviceId . "\",\"hostname\":\"" . $hn . "\",\"kind\":\"mikrotik\",\"role\":\"edge\"}")

  :if ([:len $NdToken] = 0) do={
    :local eurl ($NdServer . "/api/edge/enroll")
    /tool fetch url=$eurl http-method=post \
      http-header-field=("Authorization: Bearer " . $NdBootstrap . ",Content-Type: application/json") \
      http-data=$body output=user as-value
    :put "enroll sent — wait for operator approve, then set NdToken"
    :return
  }

  :local hurl ($NdServer . "/api/edge/heartbeat")
  :local href [/tool fetch url=$hurl http-method=post \
    http-header-field=("Authorization: Bearer " . $NdToken . ",Content-Type: application/json") \
    http-data=$body output=user as-value]
  # desired_hostname in JSON — operator rename (manual parse if needed)

  :local curl ($NdServer . "/api/edge/commands?device_id=" . $NdDeviceId)
  :local cres [/tool fetch url=$curl http-header-field=("Authorization: Bearer " . $NdToken) output=user as-value]
  # commands may include apply_rsc with arg=filename →
  # /tool fetch url=($NdServer."/api/edge/rsc?name=".$fname) ... dst-path=nd-apply.rsc
  # /import nd-apply.rsc
}

/system script remove [find where name="nd-agent-tick"]
/system script add name=nd-agent-tick owner=admin policy=read,write,policy,test source={
  :global NdAgentTick
  \$NdAgentTick
}
/system scheduler remove [find where name="nd-agent"]
/system scheduler add name=nd-agent interval=2m on-event="/system script run nd-agent-tick" policy=read,write,policy,test
:put "nd-agent scheduler installed"
