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

      def port
        config.fetch('port', URI::Postgres::DEFAULT_PORT)
      end

      def name
        config.fetch('name')
      end

      def version
        config.fetch('version')
      end

      def backup_manager
        config.fetch('backup_manager')
      end

      def password
        config.fetch('password')
      end

      def backup_manager
        config.fetch('backup_manager')
      end

      def params
        config.fetch('params', {})
      end

      def url
        URI::Postgres.build(components(host, port))
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

      def port
        config.fetch('port', 443)
      end

      def access_key
        config.fetch('access_key')
      end

      def secret_key
        config.fetch('secret_key')
      end

      def url
        build_uri(host: host, port: port)
      end

      def ssl?
        !!config.fetch('use_ssl')
      end

      private

      def build_uri(components)
        if ssl?
          URI::HTTPS.build(components)
        else
          URI::HTTP.build(components)
        end
      end
    end
  end
end
