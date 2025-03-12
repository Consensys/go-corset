(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (X :i16) (Y :i16@loob))

(defconstraint test ()
  (- X
     (if Y
         (vanishes! 0)
         (vanishes! 16))))
