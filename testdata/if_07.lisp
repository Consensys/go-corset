(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (X :i16@loob) (Y :i16@loob) (Z :i16))
(defconstraint test ()
  (if X
      (vanishes! 0)
      (vanishes! (- Z (if Y 3 16)))))
