#include <iostream>
#include <thread>
#include <chrono>
#include <random>
#include <map>
#include <atomic>
#include <Windows.h>

#include "resource.h"

using namespace std;

// Dictionary to map keys to their virtual key codes
map<string, int> keyToVirtualCode = {
    {"w", 0x57},    // Virtual key code for 'W'
    {"a", 0x41},    // Virtual key code for 'A'
    {"s", 0x53},    // Virtual key code for 'S'
    {"d", 0x44},    // Virtual key code for 'D'
    {"space", VK_SPACE} // Virtual key code for 'Space'
};

atomic<bool> running(false);
atomic<bool> stopRequested(false);

// Function to simulate a key press
void simulateKeyPress(int virtualKeyCode) {
    INPUT input = {0};
    input.type = INPUT_KEYBOARD;
    input.ki.wVk = virtualKeyCode;

    // Press the key
    SendInput(1, &input, sizeof(INPUT));

    // Release the key
    input.ki.dwFlags = KEYEVENTF_KEYUP;
    SendInput(1, &input, sizeof(INPUT));
}

// Toggle the script on/off
void toggleScript() {
    running = !running;
    cout << (running ? "Script Started" : "Script Stopped") << endl;
}

// Stop the script completely
void stopScript() {
    running = false;
    stopRequested = true;
    cout << "Script Stopped" << endl;
}

// Function to press a random key
void randomKeypress() {
    random_device rd;
    mt19937 gen(rd());
    uniform_int_distribution<> dist(0, keyToVirtualCode.size() - 1);
    auto it = next(keyToVirtualCode.begin(), dist(gen));

    simulateKeyPress(it->second); // Simulate key press
    this_thread::sleep_for(chrono::milliseconds(1000 + dist(gen) % 1000));
}

// Function to simulate a double key press
void doubleKeypress(const string &key) {
    auto it = keyToVirtualCode.find(key);
    if (it != keyToVirtualCode.end()) {
        for (int i = 0; i < 2; ++i) {
            simulateKeyPress(it->second);
            this_thread::sleep_for(chrono::milliseconds(50));
        }
    }
}

// AFK mode 1
void codeOption1() {
    cout << "AFK mode for LEGO activated" << endl;
    try {
        while (!stopRequested) {
            if (running) {
                randomKeypress();
            }
            this_thread::sleep_for(chrono::milliseconds(100));
        }
    } catch (...) {
        cout << "An error occurred." << endl;
        running = false;
    }
}

// AFK mode 2
void codeOption2() {
    cout << "AFK mode for AFK maps activated" << endl;
    try {
        while (!stopRequested) {
            if (running) {
                doubleKeypress("w");
                doubleKeypress("s");
                doubleKeypress("a");
                doubleKeypress("d");
                this_thread::sleep_for(chrono::seconds(180));
            } else {
                this_thread::sleep_for(chrono::milliseconds(100));
            }
        }
    } catch (...) {
        cout << "An error occurred." << endl;
        running = false;
    }
}

// Key listener thread
void keyListener() {
    while (!stopRequested) {
        if (GetAsyncKeyState(VK_OEM_PLUS) & 0x8000) { // '=' key
            toggleScript();
            this_thread::sleep_for(chrono::milliseconds(300)); // Debounce delay
        }
        if (GetAsyncKeyState(VK_OEM_MINUS) & 0x8000) { // '-' key
            stopScript();
            break;
        }
        this_thread::sleep_for(chrono::milliseconds(50)); // Small delay to reduce CPU usage
    }
}

// Main menu
void mainMenu() {
    cout << "*******************************" << endl;
    cout << "*                             *" << endl;
    cout << "*     SELECT AFK OPTION       *" << endl;
    cout << "*                             *" << endl;
    cout << "*******************************" << endl;
    cout << "Choose which AFK mode you want to use:" << endl;
    cout << "[1] - AFK #1: AFK mode for LEGO" << endl;
    cout << "[2] - AFK #2: AFK mode for AFK maps" << endl;

    int choice;
    cin >> choice;

    thread listenerThread(keyListener);

    if (choice == 1) {
        codeOption1();
    } else if (choice == 2) {
        codeOption2();
    } else {
        cout << "Invalid choice. Please try again." << endl;
        stopRequested = true;
    }

    listenerThread.join();
}

int main() {
    mainMenu();
    return 0;
}
