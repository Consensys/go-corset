(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (X :i16@loob) (Y :i16@loob) (Z :i16))
(defconstraint test ()
  (if X (vanishes! (- Z (if Y 0)))))

(defconstraint test ()
  (if X (vanishes! (- Z (if Y 0 16)))))
