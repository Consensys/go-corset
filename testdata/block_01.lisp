(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns X Y)
(defconstraint c1 ()
  (begin
   (vanishes! X)
   (vanishes! Y)))
