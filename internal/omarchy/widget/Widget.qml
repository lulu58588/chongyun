import QtQuick
import Quickshell
import qs.Commons
import qs.Ui

// 星鸣's icon in Omarchy's bar, put there by 星鸣 (Settings → Bar icon).
// A click drops down its panel under the bar, where the click was; a right
// click opens its window. 星鸣 is started if it isn't running.
BarWidget {
  id: root
  moduleName: "starling.gateway"

  // written in by 星鸣 when it puts the widget here
  readonly property string starling: "__STARLING__"

  implicitWidth: button.implicitWidth
  implicitHeight: button.implicitHeight

  BarIconButton {
    id: button
    anchors.fill: parent
    bar: root.bar
    text: String.fromCodePoint(0xf15c6) // nf-md-bird
    tooltipText: "星鸣"
    onPressed: function(b) {
      Quickshell.execDetached(b === Qt.RightButton ? [root.starling, "app"] : [root.starling, "panel"])
    }
  }
}
