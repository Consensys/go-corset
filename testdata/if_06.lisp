(defpurefun ((vanishes! :@loob) x) x)

(defcolumns X (Y :@loob))

(defconstraint test1 ()
  (- X
     (if Y
         (vanishes! 0))))

(defconstraint test2 ()
  (- X
     (if Y
         (vanishes! 0)
         (vanishes! 16))))
