(defpurefun ((vanishes! :@loob) x) x)

(defcolumns X Y)
(defconstraint c1 ()
  (begin
   (vanishes! X)
   (vanishes! Y)))
