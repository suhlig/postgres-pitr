require_relative '../uri/postgres'
require 'yaml'
require 'pathname'
require 'securerandom'

module PITR
  module Config
    class Base
      attr_reader :config

      def initialize(path, key)
        @config = YAML.load_file(path).fetch(key)
      end
    end

    class DB < Base
      def user
        config.fetch('user')
      end

      def host
        config.fetch('host')
      end

      def local_port
        config.fetch('local_port', port)
      end

      def port
        config.fetch('port', URI::Postgres::DEFAULT_PORT)
      end

      def name
        config.fetch('name')
      end

      def version
        config.fetch('version')
      end

      def password
        config.fetch('password')
      end

      def params
        config.fetch('params', {})
      end

      def url
        URI::Postgres.build( components(host, port) )
      end

      def local_url
        URI::Postgres.build( components('localhost', local_port) )
      end

      private

      def components(host, port)
        {
          userinfo: [user, password].join(':'),
          host: host,
          port: port,
          path: '/' + name,
          query: query_string,
        }
      end

      def query_string
        return if params.empty?
        params&.map{|kv| kv.join('=') }&.join('&')
      end
    end

    class Blobstore < Base
      def host
        config.fetch('host')
      end

      def local_port
        config.fetch('local_port')
      end

      def port
        config.fetch('port', 443)
      end

      def access_key
        config.fetch('access_key')
      end

      def secret_key
        config.fetch('secret_key')
      end
    end
  end
end
