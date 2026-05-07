```bash
                   _                  _   _           ___ ____  
  /\/\   __ _ _ __(_) ___  _ __   ___| |_| |_ ___    / __\___ \ 
 /    \ / _` | '__| |/ _ \| '_ \ / _ \ __| __/ _ \  / /    __) |
/ /\/\ \ (_| | |  | | (_) | | | |  __/ |_| ||  __/_/ /___ / __/ 
\/    \/\__,_|_|  |_|\___/|_| |_|\___|\__|\__\___(_)____/|_____|
```
# Marionette.C2
Docker Swarm based C2 infrastructure for red teaming.

## Abstract

Most C2 frameworks treat their own infrastructure as an afterthought. You get one teamserver, a couple of hand-rolled redirectors, and a Slack thread full of IPs to remember. When something falls over mid-op, you're rebuilding by hand at 2am. Marionette runs the whole stack as Docker Swarm services. Listeners, the teamserver, the database, and the operator dashboard are all services on segmented networks. A listener can come up or down without taking the rest of the infrastructure with it. If one gets burned, you scale it to zero and bring up a fresh one on a different transport.It's built for lab work, training environments, and authorized red team engagements. The naming is theatrical because the metaphor actually fits how an op runs. A puppeteer pulls strings on dolls across a stage. Scenes are campaigns. Whispers are the quiet channels. It's also more fun than `agent_handler_v2`. It totally wasn't a cool name i decided to base the entire C2 around lol.

I'm not claiming the world needs another C2. Sliver, Mythic, and Havoc all exist and are excellent. Marionette is a research project for me to learn distributed C2 design properly, with the swarm-native architecture as the actual contribution.

## Shorthand
- Dolls -> Hosts / agents / implants
- Strings -> Commands / taskings issued to Dolls
- Puppeteer -> Operator / controller
- Stage -> Target environment or engagement scope
- Theater -> Swarm cluster / full deployment environment
- Scenes -> Campaigns, operations, or task groups
- Props -> Payloads, tools, or uploaded files
- Whispers -> Low-noise / covert communications channel
- Cue -> Triggered action / automation event

## Roadmap
- [ ] Operator Dashboard
- [ ] Distributed listeners (HTTP/HTTPS/TCP)
- [ ] Task queue for agent commanding
- [ ] Agent registration and beacon handling
- [ ] Swarm service deployment for C2 modules
- [ ] Role-based access control for operators
- [ ] Encrypted task/result transport
- [ ] Campaign / host grouping (“Scenes”)
- [ ] File transfer / payload staging ("Props")
- [ ] Audit logging and operator action history
- [ ] Multi-network segmentation (listener net / operator net / storage net)
- [ ] Blue-team-safe lab mode / simulation mode

## Architecture

Four Go binaries, a shared library, a React dashboard, and a Docker Swarm stack that wires them together.

### Components

#### Teamserver
The brain. It owns the database, the task queue, the theater controller that talks to the swarm, and the operator-facing API. Anything that needs to persist or coordinate goes through here.

#### puppet
The operator client. CLI for now, authenticated against teamserver. This is what you actually drive an op from when you don't want a browser.

#### doll
The agent. It runs on a target host, beacons home on the schedule the puppeteer set, picks up strings, runs them, and ships results back.

#### whisper
The covert relay. The current design is a DNS-style channel, hence the base32-DNS encoder in `shared/util`, but the relay layer is general enough to host other low-noise transports later.

#### shared
Holds the parts everyone needs. `crypto` has AES-256-GCM and X25519. `util` has DNS-safe base32 and a chunker for transports with small frames. `protocol` defines the doll-to-teamserver wire format (see `shared/protocol/README.md` for the design).

#### dashboard
A React app talking to the teamserver API. Operators who don't want to live in puppet's terminal use this instead.

### How it fits together

A doll on a compromised host beacons to a listener. The listener is a swarm service running in the listener network. It peels off the transport layer and forwards the inner ciphertext to teamserver. Teamserver decrypts, queues whatever the doll sent (output, file chunk, beacon heartbeat), and hands back any pending strings on the way out.

Operators sit on the operator network, which can't reach the listener network directly. Puppet and the dashboard both hit teamserver's API. Teamserver is the only thing bridging the two sides.

The whole stack comes up as a swarm deployment. Scaling listeners, killing a burned one, or rotating transports is a `docker service` command instead of a manual rebuild.

### Network segmentation

Three networks by design:

- Listener net carries beacon traffic from dolls to listeners
- Operator net carries operator API calls and dashboard traffic
- Storage net is where the database and persistent state live, reachable only from teamserver

If something goes wrong on the listener side, the blast radius stops at that boundary.

## Anti-RE for dolls

Dolls run on hosts that don't belong to us. The doll binary should be annoying to reverse, the whisper relay similarly. Teamserver and puppet stay clean because they sit on operator infrastructure.

`make release` builds the hardened path:

- `garble` for name and string obfuscation
- `-trimpath` to drop filesystem paths from the binary
- `-ldflags "-s -w"` to strip the symbol table and DWARF info

`make build` stays unhardened for development. Switch to `make release` when packaging anything that's leaving the lab.

This is the baseline. Per-build key injection, hypervisor checks, and timing-based debugger detection are on the list.

