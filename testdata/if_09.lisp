(defpurefun ((vanishes! :@loob) x) x)

(defcolumns (X :@loob) (Y :@loob) Z)
(defconstraint test ()
  (vanishes! (- Z (if X (if Y 0 16)))))
