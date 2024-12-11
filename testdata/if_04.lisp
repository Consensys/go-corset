(defpurefun ((vanishes! :@loob) x) x)

(defcolumns X (Y :@loob))

(defconstraint test ()
  (- X
     (if Y
         (vanishes! 0)
         (vanishes! 16))))
