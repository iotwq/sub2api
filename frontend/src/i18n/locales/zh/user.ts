import mcp from './mcp'
import landing from './landing'
import common from './common'
import dashboard from './dashboard'
import channelMonitorV2 from './channelMonitorV2'
import batchImage from './batchImage'
import overview from './admin/overview'
import channels from './admin/channels'
import resources from './admin/resources'
import shared from './admin/shared'
import misc from './misc'

export default {
  mcp,
  ...landing,
  ...common,
  ...dashboard,
  ...channelMonitorV2,
  ...batchImage,
  admin: { ...overview, ...channels, ...resources, ...shared },
  ...misc,
}
