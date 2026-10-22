import QtQuick
import Quickshell
import qs.Commons
import qs.Ui

// 崇云科技's icon in Omarchy's bar, put there by 崇云科技 (Settings → Bar icon).
// A click drops down its panel under the bar, where the click was; a right
// click opens its window. 崇云科技 is started if it isn't running.
BarWidget {
  id: root
  moduleName: "chongyun.gateway"

  // written in by 崇云科技 when it puts the widget here
  readonly property string chongyun: "__CHONGYUN__"

  implicitWidth: button.implicitWidth
  implicitHeight: button.implicitHeight

  BarIconButton {
    id: button
    anchors.fill: parent
    bar: root.bar
    text: String.fromCodePoint(0xf15c6) // nf-md-bird
    tooltipText: "崇云科技"
    onPressed: function(b) {
      Quickshell.execDetached(b === Qt.RightButton ? [root.chongyun, "app"] : [root.chongyun, "panel"])
    }
  }
}
