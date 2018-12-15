require_relative '../uri/postgres'
require 'yaml'

module PITR
  module Config
    class DB
      def initialize(path)
        @config = YAML.load_file(path)['db']
      end

      def user
        @config['user']
      end

      def password
        File.read('ansible/.postgres-password').chomp
      end

      def host
        @config.fetch('host', 'localhost')
      end

      def port
        @config.fetch('port', URI::Postgres::DEFAULT_PORT)
      end

      def name
        @config['name']
      end

      def params
        @config['params']
      end

      def url
        URI::Postgres.build(
          userinfo: [user, password].join(':'),
          host: host,
          port: port,
          path: '/' + name,
          query: params&.map{|kv| kv.join('=') }&.join('&'),
        )
      end
    end
  end
end
