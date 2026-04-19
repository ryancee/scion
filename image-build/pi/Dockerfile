# syntax=docker/dockerfile:1
# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ARG BASE_IMAGE
FROM ${BASE_IMAGE}

RUN mkdir -p /home/scion/.pi/agent && \
  chown -R scion:scion /home/scion/.pi

# Install pi coding agent
RUN npm install -g @mariozechner/pi-coding-agent \
  && npm cache clean --force

# Clone pi-skills and pre-install JS dependencies for skills that require it
RUN git clone --depth=1 https://github.com/badlogic/pi-skills /home/scion/.pi/agent/skills/pi-skills && \
  cd /home/scion/.pi/agent/skills/pi-skills/brave-search && bun install && \
  cd /home/scion/.pi/agent/skills/pi-skills/browser-tools && bun install && \
  cd /home/scion/.pi/agent/skills/pi-skills/youtube-transcript && bun install && \
  chown -R scion:scion /home/scion/.pi

CMD ["pi"]
