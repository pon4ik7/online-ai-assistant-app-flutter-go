import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'dart:async';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'AI Консультант',
      theme: ThemeData(
        primarySwatch: Colors.blue,
        visualDensity: VisualDensity.adaptivePlatformDensity,
      ),
      home: const ChatPage(),
      debugShowCheckedModeBanner: false,
    );
  }
}

class ChatPage extends StatefulWidget {
  const ChatPage({super.key});

  @override
  State<ChatPage> createState() => _ChatPageState();
}

class _ChatPageState extends State<ChatPage> {
  final TextEditingController _messageController = TextEditingController();
  final ScrollController _scrollController = ScrollController();
  final List<ChatMessage> _messages = [];
  bool _isLoading = false;
  bool _sessionStarted = false;
  Timer? _sessionTimer;
  String? _sessionId;
  final _client = http.Client();
  final _cookieJar = <String, String>{};

  @override
  void initState() {
    super.initState();
    _loadSession();
    _messages.add(
      const ChatMessage(
        text: 'Нажмите "Начать диалог" для начала общения с AI-консультантом',
        isUser: false,
      ),
    );
  }

  Future<void> _loadSession() async {
    final prefs = await SharedPreferences.getInstance();
    setState(() {
      _sessionId = prefs.getString('session_id');
      if (_sessionId != null) {
        _sessionStarted = true;
        _startSessionTimer();
      }
    });
  }

  Future<void> _saveSession(String sessionId) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('session_id', sessionId);
  }

  Future<void> _clearSession() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('session_id');
  }

  @override
  void dispose() {
    _messageController.dispose();
    _scrollController.dispose();
    _sessionTimer?.cancel();
    _client.close();
    super.dispose();
  }

  Future<http.Response> _postWithCookies(
    String url,
    Map<String, dynamic> body,
  ) async {
    final headers = {
      'Content-Type': 'application/json; charset=utf-8',
      'Cookie': _cookieJar.entries.map((e) => '${e.key}=${e.value}').join('; '),
    };

    final response = await _client.post(
      Uri.parse(url),
      headers: headers,
      body: jsonEncode(body),
    );

    // Сохраняем cookies из ответа
    _updateCookies(response.headers['set-cookie']);

    return response;
  }

  void _updateCookies(String? cookieHeader) {
    if (cookieHeader == null) return;

    final cookies = cookieHeader.split(';');
    for (var cookie in cookies) {
      final parts = cookie.split('=');
      if (parts.length == 2) {
        _cookieJar[parts[0].trim()] = parts[1].trim();
        if (parts[0].trim() == 'session_id') {
          _sessionId = parts[1].trim();
          _saveSession(_sessionId!);
        }
      }
    }
  }

  Future<void> _startSession() async {
    setState(() {
      _isLoading = true;
    });

    try {
      final response = await _postWithCookies(
        'http://localhost:8080/api/start',
        {},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(utf8.decode(response.bodyBytes));
        setState(() {
          _messages.clear();
          _messages.add(ChatMessage(text: data['message'], isUser: false));
          _sessionStarted = true;
          _isLoading = false;
        });
        _startSessionTimer();
      } else {
        throw Exception('Ошибка сервера: ${response.statusCode}');
      }
    } catch (e) {
      setState(() {
        _messages.add(ChatMessage(text: 'Ошибка: $e', isUser: false));
        _isLoading = false;
      });
    }
    _scrollToBottom();
  }

  Future<void> _sendMessage() async {
    if (_messageController.text.isEmpty || !_sessionStarted) return;

    final message = _messageController.text;
    setState(() {
      _messages.add(ChatMessage(text: message, isUser: true));
      _messageController.clear();
      _isLoading = true;
    });
    _scrollToBottom();

    try {
      final response = await _postWithCookies(
        'http://localhost:8080/api/message',
        {'message': message},
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(utf8.decode(response.bodyBytes));
        setState(() {
          _messages.add(ChatMessage(text: data['response'], isUser: false));
          _isLoading = false;
        });
        _resetSessionTimer();
      } else {
        throw Exception('Ошибка сервера: ${response.statusCode}');
      }
    } catch (e) {
      setState(() {
        _messages.add(ChatMessage(text: 'Ошибка: $e', isUser: false));
        _isLoading = false;
      });
    }
    _scrollToBottom();
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scrollController.hasClients) {
        _scrollController.animateTo(
          _scrollController.position.maxScrollExtent,
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeOut,
        );
      }
    });
  }

  void _startSessionTimer() {
    _sessionTimer?.cancel();
    _sessionTimer = Timer(const Duration(minutes: 15), () {
      setState(() {
        _sessionStarted = false;
        _messages.add(
          const ChatMessage(
            text: 'Сессия завершена из-за неактивности. Начните новый диалог.',
            isUser: false,
          ),
        );
        _clearSession();
        _cookieJar.clear();
      });
    });
  }

  void _resetSessionTimer() {
    _startSessionTimer();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('AI Консультант'),
        actions: [
          if (_sessionStarted)
            IconButton(
              icon: const Icon(Icons.refresh),
              onPressed: () {
                setState(() {
                  _sessionStarted = false;
                  _messages.clear();
                  _messages.add(
                    const ChatMessage(
                      text: 'Сессия сброшена. Начните новый диалог.',
                      isUser: false,
                    ),
                  );
                  _clearSession();
                  _cookieJar.clear();
                });
              },
            ),
        ],
      ),
      body: Column(
        children: [
          Expanded(
            child: ListView.builder(
              controller: _scrollController,
              itemCount: _messages.length,
              itemBuilder: (context, index) {
                return _messages[index];
              },
            ),
          ),
          if (_isLoading)
            const Padding(
              padding: EdgeInsets.all(8.0),
              child: LinearProgressIndicator(),
            ),
          Padding(
            padding: const EdgeInsets.all(8.0),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _messageController,
                    decoration: InputDecoration(
                      hintText: _sessionStarted
                          ? 'Введите сообщение...'
                          : 'Начните диалог',
                      border: const OutlineInputBorder(),
                      suffixIcon: IconButton(
                        icon: const Icon(Icons.send),
                        onPressed: _sessionStarted ? _sendMessage : null,
                      ),
                    ),
                    onSubmitted: _sessionStarted ? (_) => _sendMessage() : null,
                    enabled: _sessionStarted,
                  ),
                ),
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(8.0),
            child: SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: _isLoading ? null : _startSession,
                child: const Text('Начать диалог'),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class ChatMessage extends StatelessWidget {
  final String text;
  final bool isUser;

  const ChatMessage({super.key, required this.text, required this.isUser});

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.symmetric(vertical: 4.0, horizontal: 8.0),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (isUser) const Spacer(),
          Container(
            constraints: BoxConstraints(
              maxWidth: MediaQuery.of(context).size.width * 0.8,
            ),
            padding: const EdgeInsets.all(12.0),
            decoration: BoxDecoration(
              color: isUser
                  ? Theme.of(context).primaryColor.withOpacity(0.1)
                  : Colors.grey[200],
              borderRadius: BorderRadius.circular(12.0),
            ),
            child: Text(
              text,
              style: TextStyle(color: isUser ? Colors.blue[800] : Colors.black),
            ),
          ),
          if (!isUser) const Spacer(),
        ],
      ),
    );
  }
}
