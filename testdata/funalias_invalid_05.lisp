;;error:3:13-14:symbol already exists
(defpurefun (or x y) (* x y))
(defunalias + or)

(defcolumns X Y)
