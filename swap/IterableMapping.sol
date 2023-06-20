// SPDX-License-Identifier: MIT
pragma solidity >=0.6.6;

struct IndexValue {
    uint256 keyIndex;
    uint256 value;
}
struct KeyFlag {
    uint256 key;
    bool deleted;
}

struct ItMap {
    mapping(uint256 => IndexValue) data;
    KeyFlag[] keys;
    uint256 size;
}

library IterableMapping {
    function insert(
        ItMap storage self,
        uint256 key,
        uint256 value
    ) internal returns (bool replaced) {
        uint256 keyIndex = self.data[key].keyIndex;
        self.data[key].value = value;
        if (keyIndex > 0) return true;
        else {
            keyIndex = self.keys.length;
            self.keys.push();
            self.data[key].keyIndex = keyIndex + 1;
            self.keys[keyIndex].key = key;
            self.size++;
            return false;
        }
    }

    function remove(ItMap storage self, uint256 key) internal returns (bool success) {
        uint256 keyIndex = self.data[key].keyIndex;
        if (keyIndex == 0) return false;
        delete self.data[key];
        self.keys[keyIndex - 1].deleted = true;
        self.size--;
    }

    function contains(ItMap storage self, uint256 key) internal view returns (bool) {
        return self.data[key].keyIndex > 0;
    }

    function get(ItMap storage self, uint256 key) internal view returns (uint256 value) {
        return self.data[key].value;
    }

    function iterateStart(ItMap storage self) internal view returns (uint256 keyIndex) {
        return iteratorSkipDeleted(self, 0);
    }

    function iterateValid(ItMap storage self, uint256 keyIndex) internal view returns (bool) {
        return keyIndex < self.keys.length;
    }

    function iterateNext(ItMap storage self, uint256 keyIndex) internal view returns (uint256 r_keyIndex) {
        return iteratorSkipDeleted(self, keyIndex + 1);
    }

    function iterateGet(ItMap storage self, uint256 keyIndex) internal view returns (uint256 key, uint256 value) {
        key = self.keys[keyIndex].key;
        value = self.data[key].value;
    }

    function iteratorSkipDeleted(ItMap storage self, uint256 keyIndex) internal view returns (uint256) {
        while (keyIndex < self.keys.length && self.keys[keyIndex].deleted) keyIndex++;
        return keyIndex;
    }
}