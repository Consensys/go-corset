(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns X (Y :i16@loob))

(defconstraint test1 ()
  (- X
     (if Y
         (vanishes! 0))))

(defconstraint test2 ()
  (- X
     (if Y
         (vanishes! 0)
         (vanishes! 16))))
